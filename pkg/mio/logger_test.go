package mio

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger_Info(t *testing.T) {
	buffer := new(bytes.Buffer)
	logger := NewLogger(buffer)
	logger.Info("test message", "key1", "value1")

	assert.Equal(t, buffer.String(), infoText+"\ttest message\tkey1: value1 \n", "Info log does not match expected format")
}

func TestLogger_Warn(t *testing.T) {
	buffer := new(bytes.Buffer)
	logger := NewLogger(buffer)
	logger.Warn("warn message", "key2", "value2")

	assert.Equal(t, buffer.String(), warnText+"\twarn message\tkey2: value2 \n", "Warn log does not match expected format")
}

func TestLogger_Error(t *testing.T) {
	buffer := new(bytes.Buffer)
	logger := NewLogger(buffer)
	logger.Error("error message", "key3", "value3")

	assert.Equal(t, buffer.String(), errorText+"\terror message\tkey3: value3 \n", "Error log does not match expected format")
}

func TestLogger_Debug(t *testing.T) {
	buffer := new(bytes.Buffer)
	logger := NewLogger(buffer)
	logger.Debug("debug message", "key4", "value4")

	assert.Equal(t, buffer.String(), debugText+"\tdebug message\tkey4: value4 \n", "Debug log does not match expected format")
}

func TestLogger_Named(t *testing.T) {
	buffer := new(bytes.Buffer)
	baseLogger := NewLogger(buffer).Named("base")
	namedLogger := baseLogger.Named("named")
	namedLogger.Info("named logger message", "key5", "value5")

	expectedName := "base.named"
	assert.Contains(t, namedLogger.(*logger).name, expectedName, "Logger name does not include expected prefix")
}
