package serializable_meta

import (
	"encoding/json"

	"github.com/qor/qor"
	"github.com/qor/qor/admin"
	"github.com/qor/qor/resource"
)

type SerializableMetaInterface interface {
	GetSerializeArgumentResource() *admin.Resource
	GetSerializeArgument(SerializableMetaInterface) interface{}
	GetSerializeArgumentKind() string
	SetSerializeArgumentValue(interface{})
}

type SerializeArgument struct {
	Kind  string
	Value string `sql:"size:65532"`
}

func (serialize SerializeArgument) GetSerializeArgumentKind() string {
	return serialize.Kind
}

func (serialize *SerializeArgument) GetSerializeArgument(serializableMetaInterface SerializableMetaInterface) interface{} {
	if res := serializableMetaInterface.GetSerializeArgumentResource(); res != nil {
		value := res.NewStruct()
		json.Unmarshal([]byte(serialize.Value), value)
		return value
	}
	return nil
}

func (serialize *SerializeArgument) SetSerializeArgumentValue(value interface{}) {
	if bytes, err := json.Marshal(value); err == nil {
		serialize.Value = string(bytes)
	}
}

func (serialize *SerializeArgument) ConfigureQorResourceBeforeInitialize(res resource.Resourcer) {
	if res, ok := res.(*admin.Resource); ok {
		if _, ok := res.Value.(SerializableMetaInterface); ok {
			if res.GetMeta("Kind") == nil {
				res.Meta(&admin.Meta{
					Name: "Kind",
					Type: "hidden",
					Valuer: func(value interface{}, context *qor.Context) interface{} {
						if context.GetDB().NewScope(value).PrimaryKeyZero() {
							return nil
						} else {
							return value.(SerializableMetaInterface).GetSerializeArgumentKind()
						}
					},
				})
			}

			if res.GetMeta("SerializeArgument") == nil {
				res.Meta(&admin.Meta{
					Name: "SerializeArgument",
					Type: "serialize_argument",
					Valuer: func(value interface{}, context *qor.Context) interface{} {
						if serializeArgument, ok := value.(SerializableMetaInterface); ok {
							return struct {
								Value    interface{}
								Resource *admin.Resource
							}{
								Value:    serializeArgument.GetSerializeArgument(serializeArgument),
								Resource: serializeArgument.GetSerializeArgumentResource(),
							}
						}
						return nil
					},
					Setter: func(result interface{}, metaValue *resource.MetaValue, context *qor.Context) {
						if serializeArgument, ok := result.(SerializableMetaInterface); ok {
							serializeArgumentResource := serializeArgument.GetSerializeArgumentResource()
							value := serializeArgumentResource.NewStruct()

							for _, meta := range serializeArgumentResource.GetMetas([]string{}) {
								if metaValue := metaValue.MetaValues.Get(meta.GetName()); metaValue != nil {
									if setter := meta.GetSetter(); setter != nil {
										setter(value, metaValue, context)
									}
								}
							}

							serializeArgument.SetSerializeArgumentValue(value)
						}
					},
				})
			}

			res.NewAttrs("Kind", "SerializeArgument")
			res.EditAttrs("ID", "Kind", "SerializeArgument")
		}
	}
}
