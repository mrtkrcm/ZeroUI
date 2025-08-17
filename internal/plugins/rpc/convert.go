package rpc

import (
	"encoding/json"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Convert time.Time to protobuf timestamp
func timestampFromTime(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

// Convert protobuf timestamp to time.Time
func timeFromTimestamp(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime()
}

// Convert interface{} to protobuf Any
func convertInterfaceToAny(value interface{}) (*anypb.Any, error) {
	if value == nil {
		return nil, nil
	}
	
	// Convert to JSON first for consistent serialization
	jsonData, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal value to JSON: %w", err)
	}
	
	any := &anypb.Any{
		TypeUrl: "type.googleapis.com/google.protobuf.Value",
		Value:   jsonData,
	}
	
	return any, nil
}

// Convert protobuf Any to interface{}
func convertAnyToInterface(any *anypb.Any) (interface{}, error) {
	if any == nil {
		return nil, nil
	}
	
	var value interface{}
	err := json.Unmarshal(any.Value, &value)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal Any value: %w", err)
	}
	
	return value, nil
}

// Convert map[string]interface{} to map[string]*anypb.Any
func convertFieldsToProto(fields map[string]interface{}) (map[string]*anypb.Any, error) {
	if fields == nil {
		return nil, nil
	}
	
	protoFields := make(map[string]*anypb.Any)
	for key, value := range fields {
		anyValue, err := convertInterfaceToAny(value)
		if err != nil {
			return nil, fmt.Errorf("failed to convert field %s: %w", key, err)
		}
		protoFields[key] = anyValue
	}
	
	return protoFields, nil
}

// Convert map[string]*anypb.Any to map[string]interface{}
func convertProtoToFields(protoFields map[string]*anypb.Any) (map[string]interface{}, error) {
	if protoFields == nil {
		return nil, nil
	}
	
	fields := make(map[string]interface{})
	for key, anyValue := range protoFields {
		value, err := convertAnyToInterface(anyValue)
		if err != nil {
			return nil, fmt.Errorf("failed to convert field %s: %w", key, err)
		}
		fields[key] = value
	}
	
	return fields, nil
}

// Convert ConfigData to protobuf (no-op since we use protobuf types directly)
func convertConfigDataToProto(data *ConfigData) (*ConfigData, error) {
	return data, nil
}

// Convert protobuf to ConfigData (no-op since we use protobuf types directly) 
func convertProtoToConfigData(proto *ConfigData) (*ConfigData, error) {
	return proto, nil
}

// Convert ConfigMetadata to protobuf (no-op since we use protobuf types directly)
func convertConfigMetadataToProto(metadata *ConfigMetadata) (*ConfigMetadata, error) {
	return metadata, nil
}

// Convert protobuf to ConfigMetadata (no-op since we use protobuf types directly)
func convertProtoToConfigMetadata(proto *ConfigMetadata) (*ConfigMetadata, error) {
	return proto, nil
}