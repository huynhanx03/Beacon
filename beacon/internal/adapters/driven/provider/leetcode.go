package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"beacon/internal/core/entity"
)

const leetcodeGraphQL = "https://leetcode.com/graphql"

const dailyQuery = `{
  "query": "query questionOfToday { activeDailyCodingChallengeQuestion { link question { questionFrontendId title difficulty topicTags { name } } } }"
}`

type LeetCodeProvider struct {
	client *http.Client
}

func NewLeetCodeProvider() *LeetCodeProvider {
	return &LeetCodeProvider{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

type graphQLResponse struct {
	Data struct {
		ActiveDailyCodingChallengeQuestion struct {
			Link     string `json:"link"`
			Question struct {
				ID         string `json:"questionFrontendId"`
				Title      string `json:"title"`
				Difficulty string `json:"difficulty"`
				TopicTags  []struct {
					Name string `json:"name"`
				} `json:"topicTags"`
			} `json:"question"`
		} `json:"activeDailyCodingChallengeQuestion"`
	} `json:"data"`
}

func (p *LeetCodeProvider) FetchDaily(ctx context.Context) (entity.DailyChallenge, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, leetcodeGraphQL, strings.NewReader(dailyQuery))
	if err != nil {
		return entity.DailyChallenge{}, fmt.Errorf("create leetcode request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://leetcode.com")
	req.Header.Set("Origin", "https://leetcode.com")

	resp, err := p.client.Do(req)
	if err != nil {
		return entity.DailyChallenge{}, fmt.Errorf("fetch leetcode daily: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return entity.DailyChallenge{}, fmt.Errorf("leetcode returned status %d", resp.StatusCode)
	}

	var result graphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return entity.DailyChallenge{}, fmt.Errorf("decode leetcode response: %w", err)
	}

	q := result.Data.ActiveDailyCodingChallengeQuestion.Question
	link := result.Data.ActiveDailyCodingChallengeQuestion.Link

	var tags []string
	for _, t := range q.TopicTags {
		tags = append(tags, t.Name)
	}

	return entity.DailyChallenge{
		ID:         q.ID,
		Title:      q.Title,
		Difficulty: q.Difficulty,
		Link:       "https://leetcode.com" + link,
		TopicTags:  tags,
	}, nil
}
