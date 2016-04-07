package main

import (
	"encoding/csv"
	"errors"
	"io"
	"math/rand"
)

const (
	creatorPoints = 1000
	correctPoints = 1500
)

var (
	CompletedError = errors.New("Already completed")
	NoAnswerError  = errors.New("No such answer")
	DupAnswerError = errors.New("Answer already exists")
	OwnAnswerError = errors.New("Choose own answer")
)

type QuestionRepo struct {
	questions []*Question
	answers   []string
}

func NewQuestionRepo(r io.Reader) (*QuestionRepo, error) {
	// Format: Question,Answer
	repo := &QuestionRepo{}
	answerSet := make(map[string]struct{})
	csv := csv.NewReader(r)
	csv.FieldsPerRecord = 2
	for {
		row, err := csv.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		question := &Question{Text: row[0], Multiplier: 1}
		answer := row[1]
		question.Answers = append(question.Answers, &Answer{Text: answer, Correct: true})
		repo.questions = append(repo.questions, question)
		answerSet[answer] = struct{}{}
	}
	repo.answers = make([]string, 0, len(answerSet))
	for answer := range answerSet {
		repo.answers = append(repo.answers, answer)
	}
	return repo, nil
}

func (r *QuestionRepo) Questions(questions []*Question) []*Question {
	set := map[int]struct{}{}
	n := cap(questions) - len(questions)
	for len(set) < n && len(set) < len(r.questions) {
		set[rand.Intn(len(r.questions))] = struct{}{}
	}
	for i := range set {
		questions = append(questions, r.questions[i])
	}
	return questions
}

func (r *QuestionRepo) Answers(answers []string) []string {
	set := map[int]struct{}{}
	n := cap(answers) - len(answers)
	for len(set) < n && len(set) < len(r.answers) {
		set[rand.Intn(len(r.answers))] = struct{}{}
	}
	for i := range set {
		answers = append(answers, r.answers[i])
	}
	return answers
}

type Player interface{}

type Answer struct {
	Correct bool
	Text    string
	Player  Player
	Votes   []Player
}

func (a *Answer) HasVoted(player Player) bool {
	for _, vote := range a.Votes {
		if vote == player {
			return true
		}
	}
	return false
}

type Question struct {
	Text       string
	Multiplier int
	Answers    []*Answer
}

type Result struct {
	Player Player
	Offset int
}

type ResultSet struct {
	Points map[*Answer][]Result
}

func NewResultSet(q *Question) *ResultSet {
	r := &ResultSet{Points: make(map[*Answer][]Result)}
	for _, answer := range q.Answers {
		creatorOffset := 0
		for _, vote := range answer.Votes {
			if answer.Correct {
				r.Points[answer] = append(r.Points[answer], Result{
					Player: vote,
					Offset: correctPoints * q.Multiplier,
				})
			}
			creatorOffset += creatorPoints
		}
		if creatorOffset > 0 && answer.Player != nil {
			r.Points[answer] = append(r.Points[answer], Result{
				Player: answer.Player,
				Offset: creatorOffset * q.Multiplier,
			})
		}
	}
	return r
}

type AnswerCollector struct {
	Question  *Question
	Remaining int
}

func (c *AnswerCollector) Collect(player Player, text string) error {
	if c.Complete() {
		return CompletedError
	}
	var answer *Answer
	for _, a := range c.Question.Answers {
		if a.Player == player {
			answer = a
		}
		if a.Text == text {
			return DupAnswerError
		}
	}
	if answer != nil {
		return CompletedError
	}
	answer = &Answer{Text: text, Player: player}
	c.Question.Answers = append(c.Question.Answers, answer)
	c.Remaining -= 1
	return nil
}

func (c *AnswerCollector) Complete() bool {
	return c.Remaining <= 0
}

type VoteCollector struct {
	Question  *Question
	Remaining int
}

func (c *VoteCollector) Collect(player Player, text string) error {
	if c.Complete() {
		return CompletedError
	}
	var answer *Answer
	for _, a := range c.Question.Answers {
		if a.Text == text {
			answer = a
		}
		if a.HasVoted(player) {
			return CompletedError
		}
	}
	if answer == nil {
		return NoAnswerError
	}
	if answer.Player == player {
		return OwnAnswerError
	}
	answer.Votes = append(answer.Votes, player)
	c.Remaining -= 1
	return nil
}

func (c *VoteCollector) Complete() bool {
	return c.Remaining <= 0
}

type Game struct {
	Questions []*Question
	current   int
	players   map[Player]int
}

func NewGame(repo *QuestionRepo) *Game {
	game := &Game{
		Questions: make([]*Question, 0, 7),
		players:   make(map[Player]int),
	}
	game.Questions = repo.Questions(game.Questions)
	for i, question := range game.Questions {
		if i < 3 {
			question.Multiplier = 1
		} else if i < 6 {
			question.Multiplier = 2
		} else {
			question.Multiplier = 3
		}
	}
	return game
}
