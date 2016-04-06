package main

import (
	"bytes"
	"testing"
)

const (
	badFile         = `Apple?`
	twoQuestionFile = `Fruit?,Apple
	Vegetable?,Carrot`
)

type testPlayer struct {
	Name string
}

func TestQuestionRepo_bad_format(t *testing.T) {
	buf := bytes.NewBufferString(badFile)
	_, err := NewQuestionRepo(buf)
	if err == nil {
		t.Fatal("Expected bad format error")
	}
}

func TestQuestionRepo_Questions(t *testing.T) {
	buf := bytes.NewBufferString(twoQuestionFile)
	repo, err := NewQuestionRepo(buf)
	if err != nil {
		t.Fatal(err)
	}
	questions := repo.Questions(make([]*Question, 0, 1))
	if len(questions) != 1 {
		t.Errorf("Expected 1 question")
	}
	questions = repo.Questions(make([]*Question, 0, 2))
	if len(questions) != 2 {
		t.Errorf("Expected 2 questions")
	}
	questions = repo.Questions(make([]*Question, 0, 3))
	if len(questions) != 2 {
		t.Errorf("Expected 2 questions")
	}
}

func TestQuestionRepo_Answers(t *testing.T) {
	buf := bytes.NewBufferString(twoQuestionFile)
	repo, err := NewQuestionRepo(buf)
	if err != nil {
		t.Fatal(err)
	}
	answers := repo.Answers(make([]string, 0, 1))
	if len(answers) != 1 {
		t.Errorf("Expected 1 answer")
	}
	answers = repo.Answers(make([]string, 0, 2))
	if len(answers) != 2 {
		t.Errorf("Expected 2 answers")
	}
	answers = repo.Answers(make([]string, 0, 3))
	if len(answers) != 2 {
		t.Errorf("Expected 2 answers")
	}
}

func TestNewResultSet(t *testing.T) {
	p1 := &testPlayer{Name: "B1"}
	p2 := &testPlayer{Name: "B2"}
	a1 := &Answer{Text: "Apple", Correct: true}
	a2 := &Answer{Text: "Banana", Player: p1}
	a3 := &Answer{Text: "Carrot", Player: p2}
	question := &Question{
		Text:    "Fruit?",
		Answers: []*Answer{a1, a2, a3},
	}

	results := NewResultSet(question)
	if len(results.Points) > 0 {
		t.Errorf("Expected empty result set, got %#v", results.Points)
	}

	a1.Votes = append(a1.Votes, p1)
	a2.Votes = append(a2.Votes, p2)
	results = NewResultSet(question)
	if results.Points[a1][0].Player != p1 {
		t.Errorf("Expected p1 to get points for correct answer")
	}
	if results.Points[a2][0].Player != p1 {
		t.Errorf("Expected p1 to get points for p2's incorrect answer")
	}
}

func TestAnswerCollector_Collect(t *testing.T) {
	p1 := &testPlayer{}
	p2 := &testPlayer{}
	question := &Question{Text: "Fruit?"}
	collector := AnswerCollector{Question: question, Remaining: 2}

	if err := collector.Collect(p1, "Apple"); err != nil {
		t.Errorf("Expected success, got %s", err)
	}
	if collector.Remaining != 1 {
		t.Errorf("Expected 1 remaining answer, got %d", collector.Remaining)
	}
	if err := collector.Collect(p1, "Banana"); err != CompletedError {
		t.Errorf("Expected CompletedError, got %s", err)
	}
	if err := collector.Collect(p2, "Apple"); err != DupAnswerError {
		t.Errorf("Expected DupAnswerError, got %s", err)
	}
	if err := collector.Collect(p2, "Banana"); err != nil {
		t.Errorf("Expected success, got %s", err)
	}
	if collector.Remaining != 0 {
		t.Errorf("Expected 0 remaining answer, got %d", collector.Remaining)
	}
}

func TestAnswerCollector_Complete(t *testing.T) {
	question := &Question{Text: "Fruit?"}
	collector := AnswerCollector{Question: question, Remaining: 2}
	if collector.Complete() {
		t.Errorf("Expected not to be complete")
	}
	collector.Remaining = 0
	if !collector.Complete() {
		t.Errorf("Expected to be complete")
	}
}

func TestVoteCollector_Collect(t *testing.T) {
	p1 := &testPlayer{}
	p2 := &testPlayer{}
	question := &Question{
		Text: "Fruit?",
		Answers: []*Answer{
			{Text: "Apple"},
			{Text: "Banana"},
			{Text: "Carrot", Player: p1},
		},
	}
	collector := VoteCollector{Question: question, Remaining: 2}

	if err := collector.Collect(p1, "Carrot"); err != OwnAnswerError {
		t.Errorf("Expected OwnAnswerError, got %s", err)
	}
	if err := collector.Collect(p1, "Apple"); err != nil {
		t.Errorf("Expected success, got %s", err)
	}
	if collector.Remaining != 1 {
		t.Errorf("Expected 1 remaining answer, got %d", collector.Remaining)
	}
	if err := collector.Collect(p1, "Banana"); err != CompletedError {
		t.Errorf("Expected CompletedError, got %s", err)
	}
	if err := collector.Collect(p2, "Nonexistent"); err != NoAnswerError {
		t.Errorf("Expected NoAnswerError, got %s", err)
	}
	if err := collector.Collect(p2, "Apple"); err != nil {
		t.Errorf("Expected success, got %s", err)
	}
	if collector.Remaining != 0 {
		t.Errorf("Expected 0 remaining answer, got %d", collector.Remaining)
	}
}

func TestVoteCollector_Complete(t *testing.T) {
	question := &Question{Text: "Fruit?"}
	collector := VoteCollector{Question: question, Remaining: 2}
	if collector.Complete() {
		t.Errorf("Expected not to be complete")
	}
	collector.Remaining = 0
	if !collector.Complete() {
		t.Errorf("Expected to be complete")
	}
}
