// Code generated by ent, DO NOT EDIT.

package chatent

import (
	"time"

	"github.com/fanchunke/xgpt3/conversation/ent/chatent/message"
	"github.com/fanchunke/xgpt3/conversation/ent/chatent/session"
	"github.com/fanchunke/xgpt3/conversation/ent/schema"
)

// The init function reads all schema descriptors with runtime code
// (default values, validators, hooks and policies) and stitches it
// to their package variables.
func init() {
	messageFields := schema.Message{}.Fields()
	_ = messageFields
	// messageDescCreatedAt is the schema descriptor for created_at field.
	messageDescCreatedAt := messageFields[5].Descriptor()
	// message.DefaultCreatedAt holds the default value on creation for the created_at field.
	message.DefaultCreatedAt = messageDescCreatedAt.Default.(func() time.Time)
	sessionFields := schema.Session{}.Fields()
	_ = sessionFields
	// sessionDescStatus is the schema descriptor for status field.
	sessionDescStatus := sessionFields[1].Descriptor()
	// session.DefaultStatus holds the default value on creation for the status field.
	session.DefaultStatus = sessionDescStatus.Default.(bool)
	// sessionDescCreatedAt is the schema descriptor for created_at field.
	sessionDescCreatedAt := sessionFields[2].Descriptor()
	// session.DefaultCreatedAt holds the default value on creation for the created_at field.
	session.DefaultCreatedAt = sessionDescCreatedAt.Default.(func() time.Time)
	// sessionDescUpdatedAt is the schema descriptor for updated_at field.
	sessionDescUpdatedAt := sessionFields[3].Descriptor()
	// session.DefaultUpdatedAt holds the default value on creation for the updated_at field.
	session.DefaultUpdatedAt = sessionDescUpdatedAt.Default.(func() time.Time)
	// session.UpdateDefaultUpdatedAt holds the default value on update for the updated_at field.
	session.UpdateDefaultUpdatedAt = sessionDescUpdatedAt.UpdateDefault.(func() time.Time)
	// sessionDescDeletedAt is the schema descriptor for deleted_at field.
	sessionDescDeletedAt := sessionFields[4].Descriptor()
	// session.DefaultDeletedAt holds the default value on creation for the deleted_at field.
	session.DefaultDeletedAt = sessionDescDeletedAt.Default.(int)
}
