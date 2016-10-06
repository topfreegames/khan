// khan
// https://github.com/topfreegames/khan
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package testing

import "github.com/uber-go/zap"

//NewMockKV is a mock key value store
func NewMockKV() *MockKeyValue {
	return &MockKeyValue{
		Values: map[string]interface{}{},
	}
}

//MockKeyValue store
type MockKeyValue struct {
	Values map[string]interface{}
}

//AddBool to the kv
func (m *MockKeyValue) AddBool(key string, value bool) {
	m.Values[key] = value
}

//AddFloat64 to the kv
func (m *MockKeyValue) AddFloat64(key string, value float64) {
	m.Values[key] = value
}

//AddInt to the kv
func (m *MockKeyValue) AddInt(key string, value int) {
	m.Values[key] = value
}

//AddInt64 to the kv
func (m *MockKeyValue) AddInt64(key string, value int64) {
	m.Values[key] = value
}

//AddMarshaler to the kv
func (m *MockKeyValue) AddMarshaler(key string, marshaler zap.LogMarshaler) error {
	m.Values[key] = marshaler
	return nil
}

//AddObject to the kv
func (m *MockKeyValue) AddObject(key string, value interface{}) error {
	m.Values[key] = value
	return nil
}

//AddString to the kv
func (m *MockKeyValue) AddString(key, value string) {
	m.Values[key] = value
}

//AddUint to the kv
func (m *MockKeyValue) AddUint(key string, value uint) {
	m.Values[key] = value
}

//AddUint64 to the kv
func (m *MockKeyValue) AddUint64(key string, value uint64) {
	m.Values[key] = value
}

//AddUintptr to the kv
func (m *MockKeyValue) AddUintptr(key string, value uintptr) {
	m.Values[key] = value
}

//NewMockLogger returns a mock logger for tests
func NewMockLogger(defaultFields ...zap.Field) *MockLogger {
	return &MockLogger{
		level:         zap.DebugLevel,
		DefaultFields: defaultFields,
		Messages:      []map[string]interface{}{},
	}
}

//MockLogger implements the zap logger interface but stores all the logged messages for inspection
type MockLogger struct {
	level         zap.Level
	DefaultFields []zap.Field
	Messages      []map[string]interface{}
	Parent        *MockLogger
}

//Level returns the level
func (m *MockLogger) Level() zap.Level {
	return m.level
}

//SetLevel sets the level
func (m *MockLogger) SetLevel(level zap.Level) {
	m.level = level
}

//With returns a sub-logger
func (m *MockLogger) With(fields ...zap.Field) zap.Logger {
	return &MockLogger{
		level:         m.level,
		DefaultFields: append(m.DefaultFields, fields...),
		Messages:      m.Messages,
		Parent:        m,
	}
}

//Check created a checked message
func (m *MockLogger) Check(level zap.Level, msg string) *zap.CheckedMessage {
	logger := m
	for logger.Parent != nil {
		logger = logger.Parent
	}

	logger.Messages = append(logger.Messages, map[string]interface{}{
		"type":    "checked",
		"level":   level,
		"message": msg,
		"fields":  append([]zap.Field{}, m.DefaultFields...),
	})

	return nil
}

//StubTime stubs time
func (m *MockLogger) StubTime() {
}

//Log logs
func (m *MockLogger) Log(level zap.Level, message string, fields ...zap.Field) {
	logger := m
	for logger.Parent != nil {
		logger = logger.Parent
	}
	logger.Messages = append(logger.Messages, map[string]interface{}{
		"type":    "log",
		"level":   level,
		"message": message,
		"fields":  append(append([]zap.Field{}, m.DefaultFields...), fields...),
	})
}

//Debug debugs
func (m *MockLogger) Debug(message string, fields ...zap.Field) {
	m.Log(zap.DebugLevel, message, fields...)
}

//Info logs
func (m *MockLogger) Info(message string, fields ...zap.Field) {
	m.Log(zap.InfoLevel, message, fields...)
}

//Warn logs
func (m *MockLogger) Warn(message string, fields ...zap.Field) {
	m.Log(zap.WarnLevel, message, fields...)
}

//Error logs
func (m *MockLogger) Error(message string, fields ...zap.Field) {
	m.Log(zap.ErrorLevel, message, fields...)
}

//Fatal logs
func (m *MockLogger) Fatal(message string, fields ...zap.Field) {
	m.Log(zap.FatalLevel, message, fields...)
}

//DFatal logs
func (m *MockLogger) DFatal(message string, fields ...zap.Field) {
	m.Log(zap.FatalLevel, message, fields...)
}

//Panic logs
func (m *MockLogger) Panic(message string, fields ...zap.Field) {
	m.Log(zap.PanicLevel, message, fields...)
}

func testLogMessage(logger *MockLogger, level zap.Level, message string, fields ...interface{}) bool {
	for _, msg := range logger.Messages {
		if msg["type"] != "log" {
			continue
		}
		if msg["message"] != message {
			continue
		}
		if msg["level"] != level {
			continue
		}

		kv := NewMockKV()
		for _, field := range msg["fields"].([]zap.Field) {
			field.AddTo(kv)
		}

		found := 0
		fieldName := ""
		for index, field := range fields {
			if index%2 == 0 {
				fieldName = field.(string)
				continue
			}
			if val, ok := kv.Values[fieldName]; ok {
				if val == field {
					found++
				}
			}
		}
		if len(fields)/2 == found {
			return true
		}
	}
	return false
}
