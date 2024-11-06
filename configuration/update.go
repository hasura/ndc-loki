package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/hasura/ndc-loki/connector/client"
	"github.com/hasura/ndc-loki/connector/metadata"
	"github.com/hasura/ndc-sdk-go/utils"
	"gopkg.in/yaml.v3"
)

var nativeQueryVariableRegex = regexp.MustCompile(`\$\{?([a-zA-Z_]\w+)\}?`)

// UpdateArguments represent input arguments of the `update` command
type UpdateArguments struct {
	Dir        string `help:"The directory where the configuration.yaml file is present" short:"d" env:"HASURA_PLUGIN_CONNECTOR_CONTEXT_PATH" default:"."`
	Coroutines int    `help:"The maximum number of coroutines" short:"c" default:"5"`
}

type updateCommand struct {
	Client    *client.Client
	OutputDir string
	Config    *metadata.Configuration
}

func introspectSchema(ctx context.Context, args *UpdateArguments) error {
	start := time.Now()
	slog.Info("introspecting metadata", slog.String("dir", args.Dir))
	originalConfig, err := metadata.ReadConfiguration(args.Dir)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		originalConfig = &defaultConfiguration
	}

	apiClient, err := client.New(originalConfig.ConnectionSettings)
	if err != nil {
		return err
	}

	cmd := updateCommand{
		Client:    apiClient,
		Config:    originalConfig,
		OutputDir: args.Dir,
	}

	if err := cmd.validateNativeQueries(ctx); err != nil {
		return err
	}

	if err := cmd.writeConfigFile(); err != nil {
		return fmt.Errorf("failed to write the configuration file: %w", err)
	}

	slog.Info("introspected successfully", slog.String("exec_time", time.Since(start).Round(time.Millisecond).String()))

	return nil
}

func (uc *updateCommand) writeConfigFile() error {
	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	_, _ = writer.WriteString("# yaml-language-server: $schema=https://raw.githubusercontent.com/hasura/ndc-loki/main/jsonschema/configuration.schema.json\n")
	encoder := yaml.NewEncoder(writer)
	encoder.SetIndent(2)
	if err := encoder.Encode(uc.Config); err != nil { //nolint:all
		return fmt.Errorf("failed to encode the configuration file: %w", err)
	}
	writer.Flush()

	return os.WriteFile(uc.OutputDir+"/configuration.yaml", buf.Bytes(), 0o644)
}

func (uc *updateCommand) validateNativeQueries(ctx context.Context) error {
	if len(uc.Config.Metadata.NativeOperations.Queries) == 0 {
		return nil
	}

	newNativeQueries := make(map[string]metadata.NativeQuery)
	for key, nativeQuery := range uc.Config.Metadata.NativeOperations.Queries {
		if _, ok := uc.Config.Metadata.Models[key]; ok {
			return fmt.Errorf("duplicated native query name `%s`. That name may exist in the models collection", key)
		}
		slog.Debug(key, slog.String("type", "native_query"), slog.String("query", nativeQuery.Query))
		args, err := findNativeQueryVariables(nativeQuery)
		if err != nil {
			return fmt.Errorf("%w; query: %s", err, nativeQuery.Query)
		}
		nativeQuery.Arguments = args
		query := nativeQuery.Query

		// validate arguments and promQL syntaxes
		for k, v := range nativeQuery.Arguments {
			switch v.Type {
			case string(metadata.ScalarInt64), string(metadata.ScalarFloat64):
				query = strings.ReplaceAll(query, fmt.Sprintf("${%s}", k), "1")
			case string(metadata.ScalarString), "":
				query = strings.ReplaceAll(query, fmt.Sprintf("${%s}", k), "foo")
			case string(metadata.ScalarDuration):
				query = strings.ReplaceAll(query, fmt.Sprintf("${%s}", k), "1m")
			default:
				return fmt.Errorf("invalid argument type `%s` in the native query `%s`", k, key)
			}
		}

		err = uc.validateQuery(ctx, query)
		if err != nil {
			return fmt.Errorf("invalid native query %s: %w", key, err)
		}

		// format and replace $<name> to ${<name>}
		query, err = formatNativeQueryVariables(nativeQuery.Query, nativeQuery.Arguments)
		if err != nil {
			return err
		}

		nativeQuery.Query = query
		newNativeQueries[key] = nativeQuery
	}

	uc.Config.Metadata.NativeOperations.Queries = newNativeQueries

	return nil
}

func (uc *updateCommand) validateQuery(ctx context.Context, query string) error {
	_, err := uc.Client.FormatQuery(ctx, query)

	return err
}

var defaultConfiguration = metadata.Configuration{
	ConnectionSettings: client.ClientSettings{
		URL: utils.NewEnvStringVariable("CONNECTION_URL"),
		Headers: map[string]utils.EnvString{
			"X-Scope-OrgID": utils.NewEnvStringVariable("LOKI_ORG_ID"),
		},
	},
	Metadata: metadata.Metadata{
		Models: map[string]metadata.ModelInfo{},
		NativeOperations: metadata.NativeOperations{
			Queries: map[string]metadata.NativeQuery{},
		},
	},
	Runtime: metadata.RuntimeSettings{
		Flat:                     false,
		UnixTimeUnit:             metadata.UnixTimeNano,
		QueryConcurrencyLimit:    5,
		MutationConcurrencyLimit: 1,
		Format: metadata.RuntimeFormatSettings{
			Timestamp:   metadata.TimestampUnixNano,
			Value:       metadata.ValueFloat64,
			NaN:         "NaN",
			Inf:         "+Inf",
			NegativeInf: "-Inf",
		},
	},
}

func findNativeQueryVariables(nq metadata.NativeQuery) (map[string]metadata.NativeQueryArgumentInfo, error) {
	result := map[string]metadata.NativeQueryArgumentInfo{}
	matches := nativeQueryVariableRegex.FindAllStringSubmatchIndex(nq.Query, -1)
	if len(matches) == 0 {
		return result, nil
	}

	for _, match := range matches {
		if len(match) < 4 {
			continue
		}
		argumentInfo, name, err := evalMatchedNativeQueryVariable(nq, match)
		if err != nil {
			return nil, err
		}
		result[name] = *argumentInfo
	}

	return result, nil
}

func evalMatchedNativeQueryVariable(nq metadata.NativeQuery, match []int) (*metadata.NativeQueryArgumentInfo, string, error) {
	queryLength := len(nq.Query)
	name := nq.Query[match[2]:match[3]]
	argumentInfo := metadata.NativeQueryArgumentInfo{}

	if match[0] > 0 && nq.Query[match[0]-1] == '[' {
		// duration variables should be bounded by square brackets
		if match[1] >= queryLength || nq.Query[match[1]] != ']' {
			return nil, "", errors.New("invalid LogQL range syntax")
		}
		argumentInfo.Type = string(metadata.ScalarDuration)
	} else if match[0] > 0 {
		c := nq.Query[match[0]-1]
		if c == '"' || c == '`' {
			// duration variables should be bounded by a quote
			if match[1] >= queryLength || nq.Query[match[1]] != c {
				return nil, "", errors.New("invalid LogQL string syntax")
			}
			argumentInfo.Type = string(metadata.ScalarString)
		}
	}

	if len(nq.Arguments) > 0 {
		// merge the existing argument from the configuration file
		arg, ok := nq.Arguments[name]
		if ok {
			argumentInfo.Description = arg.Description
			if argumentInfo.Type == "" && arg.Type != "" {
				argumentInfo.Type = arg.Type
			}
		}
	}
	if argumentInfo.Type == "" {
		argumentInfo.Type = string(metadata.ScalarString)
	}

	return &argumentInfo, name, nil
}

func formatNativeQueryVariables(queryInput string, variables map[string]metadata.NativeQueryArgumentInfo) (string, error) {
	query := queryInput
	for key := range variables {
		rawPattern := fmt.Sprintf(`\$\{?%s\}?`, key)
		rg, err := regexp.Compile(rawPattern)
		if err != nil {
			return "", fmt.Errorf("failed to compile regular expression %s, query: %s, error: %w", rawPattern, queryInput, err)
		}
		query = rg.ReplaceAllLiteralString(query, fmt.Sprintf("${%s}", key))
	}

	return query, nil
}
