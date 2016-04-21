package main

import (
	"encoding/csv"
	"errors"
	"io"
	"math/rand"
	"sort"
	"strings"
)

const (
	creatorPoints = 1000
	correctPoints = 1500
)

var (
	CompletedError   = errors.New("Already completed")
	NoAnswerError    = errors.New("No such answer")
	DupAnswerError   = errors.New("Answer already exists")
	OwnAnswerError   = errors.New("Choose own answer")
	EmptyAnswerError = errors.New("Empty answer")
)

func cleanText(s string) string {
	return strings.ToUpper(strings.TrimSpace(s))
}

type record struct {
	Question string
	Answer   string
}

type QuestionRepo struct {
	records []record
	answers []string
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
		record := record{Question: row[0], Answer: cleanText(row[1])}
		repo.records = append(repo.records, record)
		answerSet[record.Answer] = struct{}{}
	}
	repo.answers = make([]string, 0, len(answerSet))
	for answer := range answerSet {
		repo.answers = append(repo.answers, answer)
	}
	return repo, nil
}

func sample(n, population int) map[int]struct{} {
	s := map[int]struct{}{}
	for len(s) < n && len(s) < population {
		s[rand.Intn(population)] = struct{}{}
	}
	return s
}

func (r *QuestionRepo) Questions(questions []*Question) []*Question {
	set := sample(cap(questions)-len(questions), len(r.records))
	for i := range set {
		record := r.records[i]
		questions = append(questions, &Question{
			Text:       record.Question,
			Multiplier: 1,
			Answers:    []*Answer{{Text: record.Answer, Correct: true}},
		})
	}
	return questions
}

func (r *QuestionRepo) Answers(answers []*Answer) []*Answer {
	set := sample(cap(answers)-len(answers), len(r.answers))
	for i := range set {
		answers = append(answers, &Answer{Text: r.answers[i]})
	}
	return answers
}

type Host interface {
	Joined(player Player)
	Question(question *Question)
	Vote(question *Question)
	Collected(player Player, complete bool)
	Results(results *ResultSet)
}

type Player interface {
	RequestAnswer(question string)
	RequestVote(question string, answers []string)
}

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

type AnswerSlice []*Answer

func (p AnswerSlice) Len() int           { return len(p) }
func (p AnswerSlice) Less(i, j int) bool { return p[i].Text < p[j].Text }
func (p AnswerSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type Question struct {
	Text       string
	Multiplier int
	Answers    AnswerSlice
}

func (q *Question) CorrectAnswer() *Answer {
	for _, answer := range q.Answers {
		if answer.Correct {
			return answer
		}
	}
	return nil
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

type NonCollector struct {
}

func (c NonCollector) Collect(player Player, text string) error {
	return CompletedError
}

func (c NonCollector) Complete() bool {
	return true
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
	text = cleanText(text)
	if text == "" {
		return EmptyAnswerError
	}
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
	sort.Sort(c.Question.Answers)
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
		if a.Text == cleanText(text) {
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

type Collector interface {
	Collect(player Player, text string) error
	Complete() bool
}

type Game struct {
	Host      Host
	Questions []*Question
	Players   map[Player]int
	current   int
	collector Collector
}

func NewGame(repo *QuestionRepo, host Host) *Game {
	game := &Game{
		Host:      host,
		Questions: make([]*Question, 0, 7),
		Players:   make(map[Player]int),
		collector: NonCollector{},
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

func (g *Game) AddPlayer(players ...Player) error {
	for _, p := range players {
		g.Players[p] = 0
		g.Host.Joined(p)
	}
	return nil
}

func (g *Game) broadcastQuestion() {
	question := g.Current()
	g.Host.Question(question)
	for player := range g.Players {
		player.RequestAnswer(question.Text)
	}
}

func (g *Game) broadcastVote() {
	question := g.Current()
	g.Host.Vote(question)
	for player := range g.Players {
		var answers []string
		for _, answer := range question.Answers {
			if answer.Player == nil || answer.Player != player {
				answers = append(answers, answer.Text)
			}
		}
		player.RequestVote(question.Text, answers)
	}
}

func (g *Game) broadcastResults(results *ResultSet) {
	g.Host.Results(results)
}

func (g *Game) complete() {
}

func (g *Game) Begin() {
	g.collector = &AnswerCollector{
		Question:  g.Current(),
		Remaining: len(g.Players),
	}
	g.broadcastQuestion()
}

func (g *Game) Vote() {
	g.collector = &VoteCollector{
		Question:  g.Current(),
		Remaining: len(g.Players),
	}
	g.broadcastVote()
}

func (g *Game) Collect(player Player, text string) error {
	err := g.collector.Collect(player, text)
	if err != nil {
		return err
	}
	g.Host.Collected(player, g.collector.Complete())
	return nil
}

func (g *Game) Stop() {
	g.collector = NonCollector{}
	results := NewResultSet(g.Current())
	for _, points := range results.Points {
		for _, offset := range points {
			g.Players[offset.Player] += offset.Offset
		}
	}
	g.broadcastResults(results)
}

func (g *Game) Current() *Question {
	return g.Questions[g.current]
}

func (g *Game) Next() {
	g.current += 1
	if g.current >= len(g.Questions) {
		g.complete()
		return
	}
	g.collector = &AnswerCollector{
		Question:  g.Current(),
		Remaining: len(g.Players),
	}
	g.broadcastQuestion()
}
