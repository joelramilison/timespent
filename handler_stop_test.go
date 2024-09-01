package main

import (
	"testing"
	"time"
)

func TestCheckDayAssignment(t *testing.T) {
	
	sessionStartTime1, err := time.Parse(time.DateTime, "2024-01-01 16:30:00")
	if err != nil {
		panic(err)
	}
	timeNow1, err := time.Parse(time.DateTime, "2024-01-01 20:30:00")
	if err != nil {
		panic(err)
	}
	sessionStartTime2, err := time.Parse(time.DateTime, "2024-01-01 16:30:00")
	if err != nil {
		panic(err)
	}
	timeNow2, err := time.Parse(time.DateTime, "2024-01-05 20:30:00")
	if err != nil {
		panic(err)
	}
	sessionStartTime3, err := time.Parse(time.DateTime, "2024-01-01 16:30:00")
	if err != nil {
		panic(err)
	}
	timeNow3, err := time.Parse(time.DateTime, "2024-01-01 22:00:00")
	if err != nil {
		panic(err)
	}
	sessionStartTime4, err := time.Parse(time.DateTime, "2024-01-01 16:30:00")
	if err != nil {
		panic(err)
	}
	timeNow4, err := time.Parse(time.DateTime, "2024-01-03 21:00:00")
	if err != nil {
		panic(err)
	}
	sessionStartTime5, err := time.Parse(time.DateTime, "2024-01-01 12:30:00")
	if err != nil {
		panic(err)
	}
	timeNow5, err := time.Parse(time.DateTime, "2024-01-01 21:00:00")
	if err != nil {
		panic(err)
	}
	sessionStartTime6, err := time.Parse(time.DateTime, "2024-01-01 19:30:00")
	if err != nil {
		panic(err)
	}
	timeNow6, err := time.Parse(time.DateTime, "2024-01-01 21:00:00")
	if err != nil {
		panic(err)
	}
	sessionStartTime7, err := time.Parse(time.DateTime, "2024-01-01 20:30:00")
	if err != nil {
		panic(err)
	}
	timeNow7, err := time.Parse(time.DateTime, "2024-01-03 21:00:00")
	if err != nil {
		panic(err)
	}

	testCases := []struct{
		name string
		clientHoursNow int
		sessionStartTime time.Time
		timeNow time.Time
		expected bool
	}{
		{
			name: "daytime",
			clientHoursNow: 10,
			sessionStartTime: sessionStartTime1,
			timeNow: timeNow1,
			expected: false,
		},
		{
			name: "daytime several days",
			clientHoursNow: 10,
			sessionStartTime: sessionStartTime2,
			timeNow: timeNow2,
			expected: false,
		},
		{
			name: "nighttime ask day",
			clientHoursNow: 10,
			sessionStartTime: sessionStartTime3,
			timeNow: timeNow3,
			expected: true,
		},
		{
			name: "nighttime several days",
			clientHoursNow: 10,
			sessionStartTime: sessionStartTime4,
			timeNow: timeNow4,
			expected: false,
		},
		{
			name: "nighttime start but afternoon end",
			clientHoursNow: 13,
			sessionStartTime: sessionStartTime5,
			timeNow: timeNow5,
			expected: false,
		},
		{
			name: "start and end nighttime same day",
			clientHoursNow: 5,
			sessionStartTime: sessionStartTime6,
			timeNow: timeNow6,
			expected: true,
		},
		{
			name: "start and end nighttime several days",
			clientHoursNow: 3,
			sessionStartTime: sessionStartTime7,
			timeNow: timeNow7,
			expected: false,
		},

	}
	for i, tc := range testCases{
		t.Run(tc.name, func(t *testing.T) {

			result := ifAskToReassignDay(tc.clientHoursNow, tc.sessionStartTime, tc.timeNow)
			if result != tc.expected {
				
				t.Errorf("Test #%v (%v) failed:\nclientHoursNow: %v\nsessionStartTime: %v\ntimeNow: %v\nexpected: %v\nactual: %v",
				i + 1, tc.name, tc.clientHoursNow, tc.sessionStartTime, tc.timeNow, tc.expected, result)
			}
		})
	}
}