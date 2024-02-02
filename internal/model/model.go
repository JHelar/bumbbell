package model

import (
	"dumbbell/internal/dto"
)

type ExerciseViewModel struct {
	Name        string
	WorkoutID   int64
	Description string
	ImageSrc    string
	WeightFrom  float64
	WeightTo    float64
	RepsFrom    float64
	RepsTo      float64
	Sets        ExerciseSetsModel
}

type ExerciseSetsModel struct {
	Items []dto.SetStatus
	Htmx  bool
}

type CardViewModel struct {
	ID          int64
	WorkoutID   int64
	ImageSrc    string
	Name        string
	Description string
}

type EditExerciseTableRowModel struct {
	IsNew       bool
	ID          int64
	SplitID     int64
	Name        string
	Description string
	WeightFrom  float64
	WeightTo    float64
	RepsFrom    float64
	RepsTo      float64
	ImageSrc    string
	Sets        int64
}

type EditWorkoutTableSplitModel struct {
	ID          int64
	Name        string
	Description string
	Exercises   []EditExerciseTableRowModel
}

type UserSettingsModel struct {
	Title  string
	Splits []EditWorkoutTableSplitModel
}

type EditExerciseModel struct {
	ID          int64
	SplitID     int64
	Name        string
	Description string
	WeightFrom  float64
	WeightTo    float64
	RepsFrom    float64
	RepsTo      float64
	ImageSrc    string
	Sets        int64
}

type EditSplitModel struct {
	ID          int64
	Name        string
	Description string
}

type DashboardPageModel struct {
	Title             string
	HasActiveWorkout  bool
	ActiveWorkout     ActiveWorkoutModel
	LatestWorkoutSets LatestWorkoutSetsModel
	WorkoutActivity   WorkoutActivityModel
	WorkoutSplits     []WorkoutSplitModel
}

type ActiveWorkoutModel struct {
	Name        string
	Description string
	Stats       ActiveWorkoutStatsModel
}

type ActiveWorkoutStatsModel struct {
	Remaining           int
	Progress            int
	Done                int
	RemainingPercentage int
	ProgressPercentage  int
	DonePercentage      int
}

type LatestWorkoutSetsModel struct {
	HasNewSet bool
	Sets      []LatestWorkoutSetModel
}

type LatestWorkoutSetModel struct {
	SplitName    string
	ExerciseName string
	Status       dto.SetStatus
}

type WorkoutActivityModel struct {
	MonthCount int
	MonthDiff  int
	Months     []WorkoutActivityMonthModel
}

type WorkoutActivityMonthModel struct {
	Month            string
	ThisYearActivity int
	LastYearActivity int
}

type WorkoutSplitModel struct {
	ID               int64
	SplitName        string
	TotalGoodRatings int
	TotalBadRatings  int
	Exercises        []WorkoutSplitExerciseModel
}

type WorkoutSplitExerciseModel struct {
	ID           int64
	ExerciseName string
	GoodRatings  int
	BadRatings   int
}

type PickExerciseModel struct {
	Title             string
	Exercises         []CardViewModel
	ActiveWorkout     ActiveWorkoutModel
	WorkoutStart      string
	WorkoutDuration   string
	WorkoutStartMilli int64
}
