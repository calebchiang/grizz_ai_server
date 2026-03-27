package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type SpeakingScores struct {
	Clarity      int `json:"clarity"`
	Articulation int `json:"articulation"`
	FillerRate   int `json:"filler_rate"`
	Pace         int `json:"pace"`
	Structure    int `json:"structure"`
}

type FillerWord struct {
	Word  string `json:"word"`
	Count int    `json:"count"`
}

type PhraseReplacement struct {
	Original    string `json:"original"`
	Replacement string `json:"replacement"`
	Reason      string `json:"reason"`
}

// INTERNAL struct used only to parse AI response
type aiDrillResponse struct {
	Clarity      int `json:"clarity"`
	Articulation int `json:"articulation"`
	FillerRate   int `json:"filler_rate"`
	Pace         int `json:"pace"`
	Structure    int `json:"structure"`

	FillerWords []FillerWord `json:"filler_words"`

	Strengths  []string `json:"strengths"`
	Weaknesses []string `json:"weaknesses"`

	PhraseReplacements []PhraseReplacement `json:"phrase_replacements"`
}

type DrillResult struct {
	Scores        SpeakingScores
	SpeakingScore int

	FillerWords []FillerWord

	Strengths  []string
	Weaknesses []string

	PhraseReplacements []PhraseReplacement
}

func buildDrillPrompt(topic string, transcript string) string {

	return fmt.Sprintf(`
You are evaluating a 60-second speaking drill.

The speaker was given the following topic:

%s

Your task is to evaluate how well they spoke about this topic.

Score the speaker from 0 to 10 in the following categories:

clarity  
How clearly the speaker communicates ideas and whether their sentences are understandable.

articulation  
How clearly the speaker pronounces words and speaks without slurring or mumbling.

filler_rate  
How frequently the speaker uses filler words such as "um", "uh", "like", "you know", etc.
10 means very few filler words.
0 means excessive filler words.

pace  
How well the speaking speed is controlled.
10 = natural pace
0 = far too fast or far too slow.

structure  
How logically organized the response is when explaining the topic.

After scoring, also provide feedback.

IMPORTANT:
Write the feedback as if you are speaking directly to the person being evaluated.

Use second-person language such as:
"You did a good job..."

DO NOT refer to them as "the user".
Always refer to them as "you".

Strengths:
Provide exactly 3 bullet points describing things the speaker did well.

Weaknesses:
Provide exactly 3 bullet points with practical speaking advice.

Phrase_replacements:
Identify phrases from the transcript that could be improved.

Return them as objects containing:
- original: the phrase from the transcript
- replacement: a stronger or clearer way to say it
- reason: a short explanation why the replacement is better

Example:

[
 {
   "original": "I was like really nervous",
   "replacement": "I felt nervous at first",
   "reason": "This removes filler language and sounds more confident."
 }
]

Provide up to 5 phrase replacements.

Filler_words:
Identify filler words used and count how many times each appears.

Example:

[
 { "word": "um", "count": 3 },
 { "word": "like", "count": 2 }
]

If no filler words exist return: []

Return ONLY valid JSON in this exact format:

{
 "clarity": number,
 "articulation": number,
 "filler_rate": number,
 "pace": number,
 "structure": number,
 "filler_words": [
   { "word": "string", "count": number }
 ],
 "strengths": [
   "string",
   "string",
   "string"
 ],
 "weaknesses": [
   "string",
   "string",
   "string"
 ],
 "phrase_replacements": [
   {
     "original": "string",
     "replacement": "string",
     "reason": "string"
   }
 ]
}

Rules:
- strengths must contain exactly 3 points
- weaknesses must contain exactly 3 points
- each point must be one sentence
- filler_words must always be an array
- if none exist return []
- return ONLY JSON

Topic:
%s

Transcript:
%s
`, topic, topic, transcript)
}

func GenerateDrillResult(topic string, transcript string) (*DrillResult, error) {

	client, err := getClient()
	if err != nil {
		return nil, err
	}

	prompt := buildDrillPrompt(topic, transcript)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:       openai.GPT4oMini,
			Temperature: 0,
			MaxTokens:   400,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a strict evaluator of public speaking and articulation.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return nil, err
	}

	content := resp.Choices[0].Message.Content
	content = strings.TrimSpace(content)

	fmt.Println("AI RAW RESPONSE:")
	fmt.Println(content)

	var aiResp aiDrillResponse

	err = json.Unmarshal([]byte(content), &aiResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI drill response: %v\nResponse: %s", err, content)
	}

	scores := SpeakingScores{
		Clarity:      aiResp.Clarity,
		Articulation: aiResp.Articulation,
		FillerRate:   aiResp.FillerRate,
		Pace:         aiResp.Pace,
		Structure:    aiResp.Structure,
	}

	total := scores.Clarity +
		scores.Articulation +
		scores.FillerRate +
		scores.Pace +
		scores.Structure

	speakingScore := int(float64(total) / 50.0 * 100.0)

	result := DrillResult{
		Scores:             scores,
		SpeakingScore:      speakingScore,
		FillerWords:        aiResp.FillerWords,
		Strengths:          aiResp.Strengths,
		Weaknesses:         aiResp.Weaknesses,
		PhraseReplacements: aiResp.PhraseReplacements,
	}

	return &result, nil
}
