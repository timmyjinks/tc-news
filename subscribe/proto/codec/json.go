package codec

import "encoding/json"

// Name is the content-subtype registered with grpc's encoding package.
const Name = "json"

// JSON is a grpc encoding.Codec that marshals messages as JSON instead of
// protobuf wire format. Used so the subscribe<->notification gRPC service
// can be hand-written without requiring protoc-generated types.
type JSON struct{}

func (JSON) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (JSON) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (JSON) Name() string {
	return Name
}
