package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad_WithValidEnvVars(t *testing.T) {
	// Clean up after test
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("CLERK_SECRET_KEY")
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("GIN_MODE")
		// Reset singleton
		configInstance = nil
	}()

	// Set valid environment variables
	os.Setenv("PORT", "3000")
	os.Setenv("CLERK_SECRET_KEY", "sk_test_12345")
	os.Setenv("DATABASE_URL", "postgresql://user:pass@localhost/db")
	os.Setenv("GIN_MODE", "release")

	config, err := Load()

	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "3000", config.Port)
	assert.Equal(t, "sk_test_12345", config.ClerkSecretKey)
	assert.Equal(t, "postgresql://user:pass@localhost/db", config.DatabaseURL)
	assert.Equal(t, "release", config.GinMode)
}

func TestLoad_WithDefaultValues(t *testing.T) {
	// Clean up after test
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("CLERK_SECRET_KEY")
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("GIN_MODE")
		// Reset singleton
		configInstance = nil
	}()

	// Set only required environment variables
	os.Setenv("CLERK_SECRET_KEY", "sk_test_12345")
	os.Setenv("DATABASE_URL", "postgresql://user:pass@localhost/db")
	// PORT and GIN_MODE not set, should use defaults

	config, err := Load()

	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "8080", config.Port, "PORT should default to 8080")
	assert.Equal(t, "debug", config.GinMode, "GIN_MODE should default to debug")
	assert.Equal(t, "sk_test_12345", config.ClerkSecretKey)
	assert.Equal(t, "postgresql://user:pass@localhost/db", config.DatabaseURL)
}

func TestLoad_MissingClerkSecretKey(t *testing.T) {
	// Clean up after test
	defer func() {
		os.Unsetenv("CLERK_SECRET_KEY")
		os.Unsetenv("DATABASE_URL")
		// Reset singleton
		configInstance = nil
	}()

	// Set only DATABASE_URL, missing CLERK_SECRET_KEY
	os.Unsetenv("CLERK_SECRET_KEY")
	os.Setenv("DATABASE_URL", "postgresql://user:pass@localhost/db")

	config, err := Load()

	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "CLERK_SECRET_KEY is required")
}

func TestLoad_MissingDatabaseURL(t *testing.T) {
	// Clean up after test
	defer func() {
		os.Unsetenv("CLERK_SECRET_KEY")
		os.Unsetenv("DATABASE_URL")
		// Reset singleton
		configInstance = nil
	}()

	// Set only CLERK_SECRET_KEY, missing DATABASE_URL
	os.Setenv("CLERK_SECRET_KEY", "sk_test_12345")
	os.Unsetenv("DATABASE_URL")

	config, err := Load()

	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "DATABASE_URL is required")
}

func TestLoad_MissingBothRequiredVars(t *testing.T) {
	// Clean up after test
	defer func() {
		os.Unsetenv("CLERK_SECRET_KEY")
		os.Unsetenv("DATABASE_URL")
		// Reset singleton
		configInstance = nil
	}()

	// Unset both required environment variables
	os.Unsetenv("CLERK_SECRET_KEY")
	os.Unsetenv("DATABASE_URL")

	config, err := Load()

	assert.Error(t, err)
	assert.Nil(t, config)
	// Should fail on CLERK_SECRET_KEY first
	assert.Contains(t, err.Error(), "CLERK_SECRET_KEY is required")
}

func TestLoad_InvalidPort(t *testing.T) {
	// Clean up after test
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("CLERK_SECRET_KEY")
		os.Unsetenv("DATABASE_URL")
		// Reset singleton
		configInstance = nil
	}()

	// Set invalid port (non-numeric)
	os.Setenv("PORT", "invalid")
	os.Setenv("CLERK_SECRET_KEY", "sk_test_12345")
	os.Setenv("DATABASE_URL", "postgresql://user:pass@localhost/db")

	config, err := Load()

	// The current implementation accepts any string as port
	// Test that it accepts the value (even if invalid)
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "invalid", config.Port)
}

func TestLoad_EmptyPort(t *testing.T) {
	// Clean up after test
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("CLERK_SECRET_KEY")
		os.Unsetenv("DATABASE_URL")
		// Reset singleton
		configInstance = nil
	}()

	// Set empty port (should default to 8080)
	os.Setenv("PORT", "")
	os.Setenv("CLERK_SECRET_KEY", "sk_test_12345")
	os.Setenv("DATABASE_URL", "postgresql://user:pass@localhost/db")

	config, err := Load()

	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "8080", config.Port, "Empty PORT should default to 8080")
}

func TestGetConfig_SingletonPattern(t *testing.T) {
	// Clean up after test
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("CLERK_SECRET_KEY")
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("GIN_MODE")
		// Reset singleton
		configInstance = nil
	}()

	// Set environment variables
	os.Setenv("PORT", "3000")
	os.Setenv("CLERK_SECRET_KEY", "sk_test_12345")
	os.Setenv("DATABASE_URL", "postgresql://user:pass@localhost/db")
	os.Setenv("GIN_MODE", "release")

	// First call
	config1, err1 := GetConfig()
	assert.NoError(t, err1)
	assert.NotNil(t, config1)

	// Second call should return same instance
	config2, err2 := GetConfig()
	assert.NoError(t, err2)
	assert.NotNil(t, config2)

	// Verify they point to the same instance
	assert.Equal(t, config1, config2, "GetConfig should return the same instance (singleton)")
	assert.True(t, config1 == config2, "GetConfig should return the same pointer")

	// Verify values match
	assert.Equal(t, "3000", config1.Port)
	assert.Equal(t, "3000", config2.Port)
}

func TestGetConfig_AfterLoad(t *testing.T) {
	// Clean up after test
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("CLERK_SECRET_KEY")
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("GIN_MODE")
		// Reset singleton
		configInstance = nil
	}()

	// Set environment variables
	os.Setenv("PORT", "4000")
	os.Setenv("CLERK_SECRET_KEY", "sk_test_abc")
	os.Setenv("DATABASE_URL", "postgresql://host/db")
	os.Setenv("GIN_MODE", "test")

	// Call Load first
	loadedConfig, err := Load()
	assert.NoError(t, err)
	assert.NotNil(t, loadedConfig)

	// GetConfig should return the same instance
	gotConfig, err := GetConfig()
	assert.NoError(t, err)
	assert.Equal(t, loadedConfig, gotConfig)
}

func TestConfigStructValues(t *testing.T) {
	// Clean up after test
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("CLERK_SECRET_KEY")
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("GIN_MODE")
		// Reset singleton
		configInstance = nil
	}()

	testCases := []struct {
		name           string
		envVars        map[string]string
		expectedConfig Config
	}{
		{
			name: "all values set",
			envVars: map[string]string{
				"PORT":             "9000",
				"CLERK_SECRET_KEY": "sk_live_key",
				"DATABASE_URL":     "postgresql://prod/db",
				"GIN_MODE":         "release",
			},
			expectedConfig: Config{
				Port:           "9000",
				ClerkSecretKey: "sk_live_key",
				DatabaseURL:    "postgresql://prod/db",
				GinMode:        "release",
			},
		},
		{
			name: "defaults applied",
			envVars: map[string]string{
				"CLERK_SECRET_KEY": "sk_test_key",
				"DATABASE_URL":     "postgresql://test/db",
			},
			expectedConfig: Config{
				Port:           "8080",
				ClerkSecretKey: "sk_test_key",
				DatabaseURL:    "postgresql://test/db",
				GinMode:        "debug",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset singleton before each sub-test
			configInstance = nil

			// Clean up before test
			os.Unsetenv("PORT")
			os.Unsetenv("CLERK_SECRET_KEY")
			os.Unsetenv("DATABASE_URL")
			os.Unsetenv("GIN_MODE")

			// Set environment variables
			for key, value := range tc.envVars {
				os.Setenv(key, value)
			}

			// Clean up after sub-test
			defer func() {
				for key := range tc.envVars {
					os.Unsetenv(key)
				}
				configInstance = nil
			}()

			config, err := Load()
			assert.NoError(t, err)
			assert.NotNil(t, config)

			// Verify all struct fields match expected values
			assert.Equal(t, tc.expectedConfig.Port, config.Port)
			assert.Equal(t, tc.expectedConfig.ClerkSecretKey, config.ClerkSecretKey)
			assert.Equal(t, tc.expectedConfig.DatabaseURL, config.DatabaseURL)
			assert.Equal(t, tc.expectedConfig.GinMode, config.GinMode)
		})
	}
}

func TestLoad_ConcurrentAccess(t *testing.T) {
	// Clean up after test
	defer func() {
		os.Unsetenv("CLERK_SECRET_KEY")
		os.Unsetenv("DATABASE_URL")
		configInstance = nil
	}()

	os.Setenv("CLERK_SECRET_KEY", "sk_test_concurrent")
	os.Setenv("DATABASE_URL", "postgresql://localhost/test")

	// Reset singleton
	configInstance = nil

	// Test that concurrent calls to Load work correctly
	done := make(chan bool, 10)
	configs := make([]*Config, 10)

	for i := 0; i < 10; i++ {
		go func(index int) {
			config, err := Load()
			if err == nil && config != nil {
				configs[index] = config
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// All should have the same config values
	for i := 0; i < 10; i++ {
		assert.NotNil(t, configs[i])
		assert.Equal(t, "sk_test_concurrent", configs[i].ClerkSecretKey)
		assert.Equal(t, "postgresql://localhost/test", configs[i].DatabaseURL)
	}
}
