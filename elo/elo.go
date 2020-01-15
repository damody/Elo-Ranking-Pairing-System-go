package elo

import (
	"sort"
	"math"
)

type EloRank struct {
	K float32
}

func Mean(number []int32) float32 {
	sum := int32(0)
	for _, i := range number {
		sum += i
	}
	return float32(float32(sum) / float32(len(number)))
}

func Median(numbers []int32) int32 {
	sort.Slice(numbers, func(i, j int) bool { return numbers[i] < numbers[j] })
	mid := len(numbers) / 2
	if (len(numbers) % 2 == 0) {
		return int32(Mean([]int32{numbers[mid-1], numbers[mid]}))
	} else {
		return numbers[mid]
	}
}

func (ER *EloRank) Get_expected(a float32, b float32) float32 {
	return float32(1.0/(1.0+math.Pow(float64(10), float64((b-a)/400))))
}

func (ER *EloRank) Rating(expected float32, actual float32, current float32) float32 {
	return float32(math.Round(float64(current + ER.K*(actual-expected))))
}

func (ER *EloRank) Compute_elo(win int32, lose int32) (int32, int32) {
	ewin := ER.Get_expected(float32(win), float32(lose))
	elose := ER.Get_expected(float32(lose), float32(win))
	rwin := ER.Rating(float32(ewin), 1.0, float32(win))
	rlose := ER.Rating(float32(elose), 1.0, float32(lose))
	return int32(rwin), int32(rlose)
}

func (ER *EloRank) Compute_elo_team(winteam []int32, loseteam []int32) ([]int32, []int32) {
	win := Mean(winteam)
	lose := Mean(loseteam)
	wint := []int32{}
	loset := []int32{}
	for _, score := range winteam {
		ewin := ER.Get_expected(float32(score), float32(lose))
		rwin := ER.Rating(float32(ewin), 1.0, float32(score))
		wint = append(wint, int32(rwin))
	}
	for _, score := range loseteam {
		elose := ER.Get_expected(float32(score), float32(win))
		rlose := ER.Rating(float32(elose), 0.0, float32(score))
		loset = append(loset, int32(rlose))
	}
	return wint, loset
}

func (ER *EloRank) Compute_elo_battle_ground(team []int32, win_mount uint, scale float32) []int32 {
	m := Mean(team)
	rest := []int32{}
	a := float32(win_mount) * scale + 0.25
	for _, score := range team {
		ewin := ER.Get_expected(float32(score), float32(m))
		rwin := ER.Rating(float32(ewin), a, float32(score))
		a -= scale
		rest = append(rest, int32(rwin))
	}
	return rest
}

