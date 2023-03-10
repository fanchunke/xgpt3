// Code generated by ent, DO NOT EDIT.

package chatent

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/fanchunke/xgpt3/conversation/ent/chatent/message"
	"github.com/fanchunke/xgpt3/conversation/ent/chatent/session"
)

// Message is the model entity for the Message schema.
type Message struct {
	config `json:"-"`
	// ID of the ent.
	ID int `json:"id,omitempty"`
	// 会话Id
	SessionID int `json:"session_id,omitempty"`
	// 消息发送者Id
	FromUserID string `json:"from_user_id,omitempty"`
	// 消息接收者Id
	ToUserID string `json:"to_user_id,omitempty"`
	// 消息内容
	Content string `json:"content,omitempty"`
	// SpouseID holds the value of the "spouse_id" field.
	SpouseID int `json:"spouse_id,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the MessageQuery when eager-loading is set.
	Edges MessageEdges `json:"edges"`
}

// MessageEdges holds the relations/edges for other nodes in the graph.
type MessageEdges struct {
	// Spouse holds the value of the spouse edge.
	Spouse *Message `json:"spouse,omitempty"`
	// Session holds the value of the session edge.
	Session *Session `json:"session,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [2]bool
}

// SpouseOrErr returns the Spouse value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e MessageEdges) SpouseOrErr() (*Message, error) {
	if e.loadedTypes[0] {
		if e.Spouse == nil {
			// Edge was loaded but was not found.
			return nil, &NotFoundError{label: message.Label}
		}
		return e.Spouse, nil
	}
	return nil, &NotLoadedError{edge: "spouse"}
}

// SessionOrErr returns the Session value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e MessageEdges) SessionOrErr() (*Session, error) {
	if e.loadedTypes[1] {
		if e.Session == nil {
			// Edge was loaded but was not found.
			return nil, &NotFoundError{label: session.Label}
		}
		return e.Session, nil
	}
	return nil, &NotLoadedError{edge: "session"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Message) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case message.FieldID, message.FieldSessionID, message.FieldSpouseID:
			values[i] = new(sql.NullInt64)
		case message.FieldFromUserID, message.FieldToUserID, message.FieldContent:
			values[i] = new(sql.NullString)
		case message.FieldCreatedAt:
			values[i] = new(sql.NullTime)
		default:
			return nil, fmt.Errorf("unexpected column %q for type Message", columns[i])
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Message fields.
func (m *Message) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case message.FieldID:
			value, ok := values[i].(*sql.NullInt64)
			if !ok {
				return fmt.Errorf("unexpected type %T for field id", value)
			}
			m.ID = int(value.Int64)
		case message.FieldSessionID:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field session_id", values[i])
			} else if value.Valid {
				m.SessionID = int(value.Int64)
			}
		case message.FieldFromUserID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field from_user_id", values[i])
			} else if value.Valid {
				m.FromUserID = value.String
			}
		case message.FieldToUserID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field to_user_id", values[i])
			} else if value.Valid {
				m.ToUserID = value.String
			}
		case message.FieldContent:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field content", values[i])
			} else if value.Valid {
				m.Content = value.String
			}
		case message.FieldSpouseID:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field spouse_id", values[i])
			} else if value.Valid {
				m.SpouseID = int(value.Int64)
			}
		case message.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				m.CreatedAt = value.Time
			}
		}
	}
	return nil
}

// QuerySpouse queries the "spouse" edge of the Message entity.
func (m *Message) QuerySpouse() *MessageQuery {
	return NewMessageClient(m.config).QuerySpouse(m)
}

// QuerySession queries the "session" edge of the Message entity.
func (m *Message) QuerySession() *SessionQuery {
	return NewMessageClient(m.config).QuerySession(m)
}

// Update returns a builder for updating this Message.
// Note that you need to call Message.Unwrap() before calling this method if this Message
// was returned from a transaction, and the transaction was committed or rolled back.
func (m *Message) Update() *MessageUpdateOne {
	return NewMessageClient(m.config).UpdateOne(m)
}

// Unwrap unwraps the Message entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (m *Message) Unwrap() *Message {
	_tx, ok := m.config.driver.(*txDriver)
	if !ok {
		panic("chatent: Message is not a transactional entity")
	}
	m.config.driver = _tx.drv
	return m
}

// String implements the fmt.Stringer.
func (m *Message) String() string {
	var builder strings.Builder
	builder.WriteString("Message(")
	builder.WriteString(fmt.Sprintf("id=%v, ", m.ID))
	builder.WriteString("session_id=")
	builder.WriteString(fmt.Sprintf("%v", m.SessionID))
	builder.WriteString(", ")
	builder.WriteString("from_user_id=")
	builder.WriteString(m.FromUserID)
	builder.WriteString(", ")
	builder.WriteString("to_user_id=")
	builder.WriteString(m.ToUserID)
	builder.WriteString(", ")
	builder.WriteString("content=")
	builder.WriteString(m.Content)
	builder.WriteString(", ")
	builder.WriteString("spouse_id=")
	builder.WriteString(fmt.Sprintf("%v", m.SpouseID))
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(m.CreatedAt.Format(time.ANSIC))
	builder.WriteByte(')')
	return builder.String()
}

// Messages is a parsable slice of Message.
type Messages []*Message

func (m Messages) config(cfg config) {
	for _i := range m {
		m[_i].config = cfg
	}
}
