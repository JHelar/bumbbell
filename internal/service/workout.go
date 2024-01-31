package service

import (
	"database/sql"
	"dumbbell/internal/dto"
	"dumbbell/internal/model"
	"dumbbell/internal/utils"
	"time"
)

var MONTH_NAMES = []string{
	"Jan",
	"Feb",
	"Mar",
	"Apr",
	"May",
	"Jun",
	"Jul",
	"Aug",
	"Sep",
	"Oct",
	"Now",
	"Dec",
}

type WorkoutService struct {
	DB *sql.DB
}

func NewWorkoutService(db *sql.DB) WorkoutService {
	return WorkoutService{DB: db}
}

func (s *WorkoutService) GetActiveWorkout(userId int64) (model.ActiveWorkoutModel, error) {
	activeWorkout, err := dto.GetActiveWorkout(userId, s.DB)

	if err != nil {
		return model.ActiveWorkoutModel{}, err
	}

	split, _ := dto.GetSplit(userId, activeWorkout.SplitID, s.DB)

	allExercises, _ := dto.GetAllExercises(activeWorkout.SplitID, s.DB)
	availableExercises, _ := dto.GetAvailableExercises(activeWorkout.SplitID, activeWorkout.ID, s.DB)
	activeWorkoutSet, _ := dto.GetActiveWorkoutSet(activeWorkout.ID, s.DB)

	all := len(allExercises)

	remaining := len(availableExercises)
	done := all - remaining

	progress := 0
	progressPercentage := 0
	if (activeWorkoutSet != dto.WorkoutSet{}) {
		progress = 1
		done -= 1
		progressPercentage = int(utils.PercentOf(int(activeWorkoutSet.SetNumber), int(activeWorkoutSet.Sets)))
	}

	return model.ActiveWorkoutModel{
		Name:        split.Name,
		Description: split.Description,
		Stats: model.ActiveWorkoutStatsModel{
			Remaining:           remaining,
			Progress:            progress,
			Done:                done,
			RemainingPercentage: int(utils.PercentOf(remaining, all)),
			ProgressPercentage:  progressPercentage,
			DonePercentage:      int(utils.PercentOf(done, all)),
		},
	}, nil
}

func (s *WorkoutService) GetLatestWorkoutSets(userId int64) (model.LatestWorkoutSetsModel, error) {
	workoutSets, err := dto.GetAllWorkoutSets(userId, 10, s.DB)
	if err != nil {
		return model.LatestWorkoutSetsModel{}, err
	}

	viewModel := model.LatestWorkoutSetsModel{
		HasNewSet: false,
		Sets:      []model.LatestWorkoutSetModel{},
	}

	for _, workoutSet := range workoutSets {
		exercise, _ := dto.GetExercise(workoutSet.ExerciseID, s.DB)
		workout, _ := dto.GetWorkout(userId, workoutSet.WorkoutID, s.DB)
		split, _ := dto.GetSplit(userId, workout.SplitID, s.DB)

		viewModel.Sets = append(viewModel.Sets, model.LatestWorkoutSetModel{
			SplitName:    split.Name,
			ExerciseName: exercise.Name,
			Status:       workoutSet.SetRating,
		})

		if workoutSet.SetRating == dto.SetCurrent {
			viewModel.HasNewSet = true
		}
	}

	return viewModel, nil
}

func (s *WorkoutService) GetWorkoutActivity(userId int64) (model.WorkoutActivityModel, error) {
	workouts, err := dto.GetAllCompletedWorkouts(userId, s.DB)
	if err != nil {
		return model.WorkoutActivityModel{}, err
	}

	months := make([]model.WorkoutActivityMonthModel, 12, 12)

	thisYear := time.Now().Year()
	lastYear := thisYear - 1

	for _, workout := range workouts {
		month := workout.StartedAt.Month()
		year := workout.StartedAt.Year()

		if year >= lastYear {
			if year == lastYear {
				months[month].LastYearActivity++
			} else {
				months[month].ThisYearActivity++
			}
		}
	}

	thisYearMonthAverage := 0.0
	lastYearMonthAverage := 0.0
	for i := range months {
		months[i].Month = MONTH_NAMES[i]
		thisYearMonthAverage += float64(months[i].ThisYearActivity)
		lastYearMonthAverage += float64(months[i].LastYearActivity)
	}

	thisYearMonthAverage = thisYearMonthAverage / float64(len(months))
	lastYearMonthAverage = lastYearMonthAverage / float64(len(months))
	monthAverageDiff := utils.Change(int(lastYearMonthAverage), int(thisYearMonthAverage))

	return model.WorkoutActivityModel{
		Months:     months,
		MonthCount: int(thisYearMonthAverage),
		MonthDiff:  int(monthAverageDiff),
	}, nil
}

func (s *WorkoutService) GetWorkoutSplits(userId int64) ([]model.WorkoutSplitModel, error) {
	splitModels := []model.WorkoutSplitModel{}
	splits, _ := dto.GetSplits(userId, s.DB)

	for _, split := range splits {
		splitModel := model.WorkoutSplitModel{
			ID:               split.ID,
			SplitName:        split.Name,
			TotalGoodRatings: 0,
			TotalBadRatings:  0,
			Exercises:        []model.WorkoutSplitExerciseModel{},
		}

		exercises, _ := dto.GetAllExercises(split.ID, s.DB)

		for _, exercise := range exercises {
			workoutSets, _ := dto.GetAllWorkoutSetsForExercise(exercise.ID, s.DB)
			splitExerciseModel := model.WorkoutSplitExerciseModel{
				ID:           exercise.ID,
				ExerciseName: exercise.Name,
				GoodRatings:  0,
				BadRatings:   0,
			}

			for _, workoutSet := range workoutSets {
				if workoutSet.SetRating == dto.SetGood {
					splitExerciseModel.GoodRatings++
				} else if workoutSet.SetRating == dto.SetBad {
					splitExerciseModel.BadRatings++
				}
			}

			splitModel.TotalGoodRatings += splitExerciseModel.GoodRatings
			splitModel.TotalBadRatings += splitExerciseModel.BadRatings
			splitModel.Exercises = append(splitModel.Exercises, splitExerciseModel)
		}

		splitModels = append(splitModels, splitModel)
	}

	return splitModels, nil
}
