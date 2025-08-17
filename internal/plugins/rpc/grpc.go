package rpc

import (
	"context"
	"fmt"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

// ConfigPluginGRPC implements the gRPC plugin interface
type ConfigPluginGRPC struct {
	plugin.Plugin
	Impl ConfigPlugin
}

// GRPCServer returns a gRPC server implementation
func (p *ConfigPluginGRPC) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	RegisterConfigPluginServer(s, &GRPCServer{Impl: p.Impl})
	return nil
}

// GRPCClient returns a gRPC client implementation
func (p *ConfigPluginGRPC) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: NewConfigPluginClient(c)}, nil
}

// GRPCServer implements the gRPC server side
type GRPCServer struct {
	UnimplementedConfigPluginServer
	Impl ConfigPlugin
}

// GetInfo implementation
func (s *GRPCServer) GetInfo(ctx context.Context, req *GetInfoRequest) (*GetInfoResponse, error) {
	info, err := s.Impl.GetInfo(ctx)
	if err != nil {
		return nil, err
	}

	return &GetInfoResponse{
		Info: &PluginInfo{
			Name:         info.Name,
			Version:      info.Version,
			Description:  info.Description,
			Author:       info.Author,
			Capabilities: info.Capabilities,
			ApiVersion:   info.ApiVersion,
			Metadata:     info.Metadata,
		},
	}, nil
}

// DetectConfig implementation
func (s *GRPCServer) DetectConfig(ctx context.Context, req *DetectConfigRequest) (*DetectConfigResponse, error) {
	config, err := s.Impl.DetectConfig(ctx)
	if err != nil {
		return nil, err
	}

	return &DetectConfigResponse{
		Config: config,
	}, nil
}

// ParseConfig implementation
func (s *GRPCServer) ParseConfig(ctx context.Context, req *ParseConfigRequest) (*ParseConfigResponse, error) {
	data, err := s.Impl.ParseConfig(ctx, req.Path)
	if err != nil {
		return nil, err
	}

	configData, err := convertConfigDataToProto(data)
	if err != nil {
		return nil, fmt.Errorf("failed to convert config data: %w", err)
	}

	return &ParseConfigResponse{
		Data: configData,
	}, nil
}

// WriteConfig implementation
func (s *GRPCServer) WriteConfig(ctx context.Context, req *WriteConfigRequest) (*WriteConfigResponse, error) {
	data, err := convertProtoToConfigData(req.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to convert proto data: %w", err)
	}

	err = s.Impl.WriteConfig(ctx, req.Path, data)
	if err != nil {
		return nil, err
	}

	return &WriteConfigResponse{}, nil
}

// ValidateField implementation
func (s *GRPCServer) ValidateField(ctx context.Context, req *ValidateFieldRequest) (*ValidateFieldResponse, error) {
	value, err := convertAnyToInterface(req.Value)
	if err != nil {
		return &ValidateFieldResponse{
			Valid: false,
			Error: fmt.Sprintf("invalid value format: %v", err),
		}, nil
	}

	err = s.Impl.ValidateField(ctx, req.Field, value)
	if err != nil {
		return &ValidateFieldResponse{
			Valid: false,
			Error: err.Error(),
		}, nil
	}

	return &ValidateFieldResponse{
		Valid: true,
	}, nil
}

// ValidateConfig implementation
func (s *GRPCServer) ValidateConfig(ctx context.Context, req *ValidateConfigRequest) (*ValidateConfigResponse, error) {
	data, err := convertProtoToConfigData(req.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to convert proto data: %w", err)
	}

	err = s.Impl.ValidateConfig(ctx, data)
	if err != nil {
		// Parse validation errors if multiple
		return &ValidateConfigResponse{
			Valid: false,
			Errors: []*ValidationError{
				{
					Field:   "config",
					Code:    "validation_failed",
					Message: err.Error(),
				},
			},
		}, nil
	}

	return &ValidateConfigResponse{
		Valid: true,
	}, nil
}

// GetSchema implementation
func (s *GRPCServer) GetSchema(ctx context.Context, req *GetSchemaRequest) (*GetSchemaResponse, error) {
	metadata, err := s.Impl.GetSchema(ctx)
	if err != nil {
		return nil, err
	}

	protoMetadata, err := convertConfigMetadataToProto(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to convert metadata: %w", err)
	}

	return &GetSchemaResponse{
		Metadata: protoMetadata,
	}, nil
}

// SupportsFeature implementation
func (s *GRPCServer) SupportsFeature(ctx context.Context, req *SupportsFeatureRequest) (*SupportsFeatureResponse, error) {
	supported, err := s.Impl.SupportsFeature(ctx, req.Feature)
	if err != nil {
		return nil, err
	}

	return &SupportsFeatureResponse{
		Supported: supported,
	}, nil
}

// GRPCClient implements the gRPC client side
type GRPCClient struct {
	client ConfigPluginClient
}

// GetInfo implementation
func (c *GRPCClient) GetInfo(ctx context.Context) (*PluginInfo, error) {
	resp, err := c.client.GetInfo(ctx, &GetInfoRequest{})
	if err != nil {
		return nil, err
	}

	return &PluginInfo{
		Name:         resp.Info.Name,
		Version:      resp.Info.Version,
		Description:  resp.Info.Description,
		Author:       resp.Info.Author,
		Capabilities: resp.Info.Capabilities,
		ApiVersion:   resp.Info.ApiVersion,
		Metadata:     resp.Info.Metadata,
	}, nil
}

// DetectConfig implementation
func (c *GRPCClient) DetectConfig(ctx context.Context) (*ConfigInfo, error) {
	resp, err := c.client.DetectConfig(ctx, &DetectConfigRequest{})
	if err != nil {
		return nil, err
	}

	return resp.Config, nil
}

// ParseConfig implementation
func (c *GRPCClient) ParseConfig(ctx context.Context, path string) (*ConfigData, error) {
	resp, err := c.client.ParseConfig(ctx, &ParseConfigRequest{Path: path})
	if err != nil {
		return nil, err
	}

	return convertProtoToConfigData(resp.Data)
}

// WriteConfig implementation
func (c *GRPCClient) WriteConfig(ctx context.Context, path string, data *ConfigData) error {
	protoData, err := convertConfigDataToProto(data)
	if err != nil {
		return fmt.Errorf("failed to convert config data: %w", err)
	}

	_, err = c.client.WriteConfig(ctx, &WriteConfigRequest{
		Path: path,
		Data: protoData,
	})
	return err
}

// ValidateField implementation
func (c *GRPCClient) ValidateField(ctx context.Context, field string, value interface{}) error {
	anyValue, err := convertInterfaceToAny(value)
	if err != nil {
		return fmt.Errorf("failed to convert value: %w", err)
	}

	resp, err := c.client.ValidateField(ctx, &ValidateFieldRequest{
		Field: field,
		Value: anyValue,
	})
	if err != nil {
		return err
	}

	if !resp.Valid {
		return fmt.Errorf("validation failed: %s", resp.Error)
	}

	return nil
}

// ValidateConfig implementation
func (c *GRPCClient) ValidateConfig(ctx context.Context, data *ConfigData) error {
	protoData, err := convertConfigDataToProto(data)
	if err != nil {
		return fmt.Errorf("failed to convert config data: %w", err)
	}

	resp, err := c.client.ValidateConfig(ctx, &ValidateConfigRequest{Data: protoData})
	if err != nil {
		return err
	}

	if !resp.Valid && len(resp.Errors) > 0 {
		return fmt.Errorf("validation failed: %s", resp.Errors[0].Message)
	}

	return nil
}

// GetSchema implementation
func (c *GRPCClient) GetSchema(ctx context.Context) (*ConfigMetadata, error) {
	resp, err := c.client.GetSchema(ctx, &GetSchemaRequest{})
	if err != nil {
		return nil, err
	}

	return convertProtoToConfigMetadata(resp.Metadata)
}

// SupportsFeature implementation
func (c *GRPCClient) SupportsFeature(ctx context.Context, feature string) (bool, error) {
	resp, err := c.client.SupportsFeature(ctx, &SupportsFeatureRequest{Feature: feature})
	if err != nil {
		return false, err
	}

	return resp.Supported, nil
}

// Ensure GRPCClient implements ConfigPlugin interface
var _ ConfigPlugin = (*GRPCClient)(nil)

// Legacy plugin implementation for backward compatibility
type ConfigPluginNetRPC struct {
	Impl ConfigPlugin
}

func (p *ConfigPluginNetRPC) Server(*plugin.MuxBroker) (interface{}, error) {
	return &NetRPCServer{Impl: p.Impl}, nil
}

func (p *ConfigPluginNetRPC) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &NetRPCClient{client: c}, nil
}

// NetRPCServer implements net/rpc server (for backward compatibility)
type NetRPCServer struct {
	Impl ConfigPlugin
}

// NetRPCClient implements net/rpc client (for backward compatibility)
type NetRPCClient struct {
	client *rpc.Client
}

// Basic implementation for legacy support - would need full implementation
func (c *NetRPCClient) GetInfo(ctx context.Context) (*PluginInfo, error) {
	var resp PluginInfo
	err := c.client.Call("Plugin.GetInfo", struct{}{}, &resp)
	return &resp, err
}
