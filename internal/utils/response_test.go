package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupGinContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

func TestSuccessResponse(t *testing.T) {
	c, w := setupGinContext()

	data := map[string]interface{}{
		"id":   1,
		"name": "Test Item",
	}

	SuccessResponse(c, data)

	assert.Equal(t, http.StatusOK, w.Code)

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response.Status)
	assert.NotNil(t, response.Data)

	// Verify data content
	dataMap, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(1), dataMap["id"])
	assert.Equal(t, "Test Item", dataMap["name"])
}

func TestSuccessResponse_WithNilData(t *testing.T) {
	c, w := setupGinContext()

	SuccessResponse(c, nil)

	assert.Equal(t, http.StatusOK, w.Code)

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response.Status)
	assert.Nil(t, response.Data)
}

func TestSuccessResponse_WithArrayData(t *testing.T) {
	c, w := setupGinContext()

	data := []string{"item1", "item2", "item3"}
	SuccessResponse(c, data)

	assert.Equal(t, http.StatusOK, w.Code)

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response.Status)
	assert.NotNil(t, response.Data)

	// Verify array data
	dataArray, ok := response.Data.([]interface{})
	assert.True(t, ok)
	assert.Len(t, dataArray, 3)
}

func TestErrorResponseClient_Code400(t *testing.T) {
	c, w := setupGinContext()

	ErrorResponseClient(c, http.StatusBadRequest, "Invalid input provided")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "error", response.Status)
	assert.Equal(t, "Invalid input provided", response.Message)
	assert.Equal(t, 400, response.Code)
}

func TestErrorResponseClient_Code404(t *testing.T) {
	c, w := setupGinContext()

	ErrorResponseClient(c, http.StatusNotFound, "Resource not found")

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "error", response.Status)
	assert.Equal(t, "Resource not found", response.Message)
	assert.Equal(t, 404, response.Code)
}

func TestErrorResponseClient_Code500(t *testing.T) {
	c, w := setupGinContext()

	ErrorResponseClient(c, http.StatusInternalServerError, "Internal server error")

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "error", response.Status)
	assert.Equal(t, "Internal server error", response.Message)
	assert.Equal(t, 500, response.Code)
}

func TestErrorResponseClient_VariousCodes(t *testing.T) {
	testCases := []struct {
		code            int
		message         string
		expectedCode    int
		expectedMessage string
	}{
		{http.StatusBadRequest, "Bad request message", 400, "Bad request message"},
		{http.StatusUnauthorized, "Unauthorized access", 401, "Unauthorized access"},
		{http.StatusForbidden, "Forbidden resource", 403, "Forbidden resource"},
		{http.StatusNotFound, "Not found message", 404, "Not found message"},
		{http.StatusConflict, "Conflict occurred", 409, "Conflict occurred"},
		{http.StatusUnprocessableEntity, "Validation failed", 422, "Validation failed"},
		{http.StatusInternalServerError, "Server error message", 500, "Server error message"},
		{http.StatusServiceUnavailable, "Service unavailable", 503, "Service unavailable"},
	}

	for _, tc := range testCases {
		t.Run(tc.message, func(t *testing.T) {
			c, w := setupGinContext()

			ErrorResponseClient(c, tc.code, tc.message)

			assert.Equal(t, tc.code, w.Code)

			var response ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "error", response.Status)
			assert.Equal(t, tc.expectedMessage, response.Message)
			assert.Equal(t, tc.expectedCode, response.Code)
		})
	}
}

func TestErrorResponseClient_EmptyMessage(t *testing.T) {
	c, w := setupGinContext()

	ErrorResponseClient(c, http.StatusBadRequest, "")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "error", response.Status)
	assert.Equal(t, "", response.Message)
	assert.Equal(t, 400, response.Code)
}

func TestWrapError(t *testing.T) {
	originalErr := assert.AnError
	wrappedErr := WrapError("test context", originalErr)

	assert.Error(t, wrappedErr)
	assert.Contains(t, wrappedErr.Error(), "test context")
	assert.Contains(t, wrappedErr.Error(), originalErr.Error())
}

func TestWrapError_NilError(t *testing.T) {
	wrappedErr := WrapError("test context", nil)

	assert.Error(t, wrappedErr)
	assert.Contains(t, wrappedErr.Error(), "test context")
	assert.Contains(t, wrappedErr.Error(), "<nil>")
}

func TestFormatErrorForClient(t *testing.T) {
	originalErr := assert.AnError
	publicMessage := "Something went wrong"

	result := FormatErrorForClient(originalErr, publicMessage)

	assert.Equal(t, publicMessage, result)
}

func TestFormatErrorForClient_NilError(t *testing.T) {
	publicMessage := "No error occurred"

	result := FormatErrorForClient(nil, publicMessage)

	assert.Equal(t, publicMessage, result)
}

func TestIsValidDate_ValidDates(t *testing.T) {
	validDates := []string{
		"2024-01-01",
		"2024-12-31",
		"2024-02-29", // Leap year
		"2000-02-29", // Leap year - divisible by 400
		"2023-12-31",
		"1999-06-15",
		"2024-06-30",
	}

	for _, date := range validDates {
		t.Run(date, func(t *testing.T) {
			assert.True(t, IsValidDate(date), "Date %s should be valid", date)
		})
	}
}

func TestIsValidDate_InvalidDates(t *testing.T) {
	invalidDates := []string{
		"2024-02-30",          // Invalid: February has 28/29 days
		"2023-02-29",          // Invalid: Not a leap year
		"2024-04-31",          // Invalid: April has 30 days
		"2024-06-31",          // Invalid: June has 30 days
		"2024-09-31",          // Invalid: September has 30 days
		"2024-11-31",          // Invalid: November has 30 days
		"2024-13-01",          // Invalid: Month 13
		"2024-00-01",          // Invalid: Month 00
		"2024-01-00",          // Invalid: Day 00
		"2024-01-32",          // Invalid: Day 32
		"01-01-2024",          // Invalid: Wrong format (DD-MM-YYYY)
		"2024/01/01",          // Invalid: Wrong separator
		"2024-1-01",           // Invalid: Single digit month
		"2024-01-1",           // Invalid: Single digit day
		"24-01-01",            // Invalid: 2-digit year
		"2024--01-01",         // Invalid: Double separator
		"2024-01-01-",         // Invalid: Extra separator at end
		"not-a-date",          // Invalid: Not a date
		"2024-01-01T00:00:00", // Invalid: DateTime format
	}

	for _, date := range invalidDates {
		t.Run(date, func(t *testing.T) {
			assert.False(t, IsValidDate(date), "Date %s should be invalid", date)
		})
	}
}

func TestIsValidDate_EmptyString(t *testing.T) {
	assert.False(t, IsValidDate(""), "Empty string should be invalid")
}

func TestIsValidDate_Whitespace(t *testing.T) {
	assert.False(t, IsValidDate(" "), "Whitespace should be invalid")
	assert.False(t, IsValidDate("  "), "Multiple spaces should be invalid")
	assert.False(t, IsValidDate("2024-01-01 "), "Date with trailing space should be invalid")
	assert.False(t, IsValidDate(" 2024-01-01"), "Date with leading space should be invalid")
}

func TestIsValidDate_BoundaryValues(t *testing.T) {
	// Test valid boundary values
	assert.True(t, IsValidDate("0001-01-01"), "Minimum valid date should be valid")
	assert.True(t, IsValidDate("9999-12-31"), "Maximum valid date should be valid")

	// Year 0000 is actually valid in Go's time parsing
	assert.True(t, IsValidDate("0000-01-01"), "Year 0000 is valid in Go time parsing")

	// 5-digit year should be invalid
	assert.False(t, IsValidDate("10000-01-01"), "5-digit year should be invalid")
}

func TestIsValidDate_LeapYearEdgeCases(t *testing.T) {
	// Valid leap years
	assert.True(t, IsValidDate("2000-02-29"), "Year 2000 is a leap year")
	assert.True(t, IsValidDate("2024-02-29"), "Year 2024 is a leap year")
	assert.True(t, IsValidDate("2020-02-29"), "Year 2020 is a leap year")
	assert.True(t, IsValidDate("2016-02-29"), "Year 2016 is a leap year")
	assert.True(t, IsValidDate("1904-02-29"), "Year 1904 is a leap year")

	// Invalid leap years (not actually leap years)
	assert.False(t, IsValidDate("1900-02-29"), "Year 1900 is NOT a leap year (divisible by 100 but not 400)")
	assert.False(t, IsValidDate("2100-02-29"), "Year 2100 is NOT a leap year (divisible by 100 but not 400)")
	assert.False(t, IsValidDate("2023-02-29"), "Year 2023 is NOT a leap year")
	assert.False(t, IsValidDate("2025-02-29"), "Year 2025 is NOT a leap year")
}

func TestIsValidDate_MonthEdgeCases(t *testing.T) {
	// Months with 31 days
	months31 := []int{1, 3, 5, 7, 8, 10, 12}
	for _, month := range months31 {
		dateStr := fmt.Sprintf("2024-%02d-31", month)
		t.Run(dateStr, func(t *testing.T) {
			assert.True(t, IsValidDate(dateStr), "Month %d should accept day 31", month)
		})
	}

	// Months with 30 days (should reject day 31)
	months30 := []int{4, 6, 9, 11}
	for _, month := range months30 {
		dateStr := fmt.Sprintf("2024-%02d-31", month)
		t.Run(dateStr, func(t *testing.T) {
			assert.False(t, IsValidDate(dateStr), "Month %d should reject day 31", month)
		})

		// But day 30 should be valid
		dateStr30 := fmt.Sprintf("2024-%02d-30", month)
		t.Run(dateStr30, func(t *testing.T) {
			assert.True(t, IsValidDate(dateStr30), "Month %d should accept day 30", month)
		})
	}
}
