package service

import (
	"database/sql"
	"dumbbell/internal/dto"
	"dumbbell/internal/model"
	"dumbbell/internal/utils"
	"errors"
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

var ErrorNoExercises = errors.New("No exercises available")

type WorkoutService struct {
	DB *sql.DB
}

func NewWorkoutService(db *sql.DB) WorkoutService {
	return WorkoutService{DB: db}
}

func (s *WorkoutService) GetActiveWorkoutData(userId int64, workoutId int64) (model.ActiveWorkoutModel, error) {
	workout, err := dto.GetWorkout(userId, workoutId, s.DB)

	if err != nil {
		return model.ActiveWorkoutModel{}, err
	}

	split, _ := dto.GetSplit(userId, workout.SplitID, s.DB)

	allExercises, _ := dto.GetAllExercises(workout.SplitID, s.DB)
	availableExercises, _ := dto.GetRemainingWorkoutExercises(workout.SplitID, workout.ID, s.DB)
	activeWorkoutSet, _ := dto.GetActiveWorkoutSet(workout.ID, s.DB)

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
		month := workout.StartedAt.Month() - 1
		year := workout.StartedAt.Year()

		if year >= lastYear {
			if year == lastYear {
				months[month].LastYearActivity++
			} else {
				months[month].ThisYearActivity++
			}
		}
	}

	for i := range months {
		months[i].Month = MONTH_NAMES[i]
	}

	thisYearMonthCount := months[time.Now().Month()-1].ThisYearActivity
	lastYearMonthCount := months[time.Now().Month()-1].LastYearActivity
	monthAverageDiff := utils.Change(int(lastYearMonthCount), int(thisYearMonthCount))

	return model.WorkoutActivityModel{
		Months:     months,
		MonthCount: int(thisYearMonthCount),
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

func (s *WorkoutService) GetAvailableExercises(splitId int64, workoutId int64) ([]model.CardViewModel, error) {
	exercises, err := dto.GetWorkoutExercises(splitId, workoutId, s.DB)
	if err != nil {
		return nil, err
	}

	cards := []model.CardViewModel{}
	for _, exercise := range exercises {
		cards = append(cards, model.CardViewModel{
			ID:          exercise.ID,
			Name:        exercise.Name,
			Description: exercise.Description,
			WorkoutID:   workoutId,
			ImageSrc:    exercise.GetImageURL(),
			Disabled:    exercise.HasWorkoutSet,
		})
	}

	return cards, nil
}

func (s *WorkoutService) GetSplitCards(userId int64) ([]model.CardViewModel, error) {
	splits, err := dto.GetSplits(userId, s.DB)
	if err != nil {
		return nil, err
	}

	cards := []model.CardViewModel{}
	for _, split := range splits {
		cards = append(cards, model.CardViewModel{
			ID:          split.ID,
			Name:        split.Name,
			Description: split.Description,
		})
	}

	return cards, nil
}

func GetWorkoutMetaData(workout dto.Workout) model.WorkoutMetadataModel {
	workoutStartString := workout.StartedAt.Format("15:04 2006-01-02")
	workoutDuration := time.Now().Sub(workout.StartedAt)
	workoutDurationString := utils.FmtDuration(workoutDuration)
	startedAt := workout.StartedAt.UnixMilli()

	return model.WorkoutMetadataModel{
		WorkoutStart:     workoutStartString,
		WorkoutDuration:  workoutDurationString,
		WorkoutStartedAt: startedAt,
	}
}

func (s *WorkoutService) GetPickExerciseModel(userId int64, workoutId int64) (model.PickExerciseModel, error) {
	workout, err := dto.GetWorkout(userId, workoutId, s.DB)
	if err != nil {
		return model.PickExerciseModel{}, err
	}

	exercises, err := s.GetAvailableExercises(workout.SplitID, workout.ID)
	if err != nil {
		return model.PickExerciseModel{}, err
	}

	hasRemainingExercises := false
	for _, exercise := range exercises {
		if !exercise.Disabled {
			hasRemainingExercises = true
			break
		}
	}
	if !hasRemainingExercises {
		return model.PickExerciseModel{}, ErrorNoExercises
	}

	activeWorkoutData, err := s.GetActiveWorkoutData(userId, workoutId)
	if err != nil {
		return model.PickExerciseModel{}, err
	}

	metadata := GetWorkoutMetaData(workout)
	return model.PickExerciseModel{
		Title:            "Dumbell",
		Exercises:        exercises,
		ActiveWorkout:    activeWorkoutData,
		WorkoutStart:     metadata.WorkoutStart,
		WorkoutDuration:  metadata.WorkoutDuration,
		WorkoutStartedAt: metadata.WorkoutStartedAt,
	}, nil
}
