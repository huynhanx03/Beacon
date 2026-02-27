package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"beacon/internal/core/entity"
)

const leetcodeGraphQL = "https://leetcode.com/graphql"
const leetcodeBase = "https://leetcode.com"

const dailyQuery = `{
  "query": "query questionOfToday { activeDailyCodingChallengeQuestion { link question { questionFrontendId title difficulty topicTags { name } } } }"
}`

const (
	maxRetries    = 3
	retryBaseWait = 2 * time.Second
)

type LeetCodeProvider struct {
	client *http.Client
}

func NewLeetCodeProvider() *LeetCodeProvider {
	jar, _ := cookiejar.New(nil)
	return &LeetCodeProvider{
		client: &http.Client{
			Timeout: 15 * time.Second,
			Jar:     jar,
		},
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

func (p *LeetCodeProvider) initSession(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, leetcodeBase, nil)
	if err != nil {
		return "", fmt.Errorf("create init request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("init session: %w", err)
	}
	defer resp.Body.Close()

	var csrfToken string
	for _, cookie := range p.client.Jar.Cookies(req.URL) {
		if cookie.Name == "csrftoken" {
			csrfToken = cookie.Value
			break
		}
	}
	return csrfToken, nil
}

func (p *LeetCodeProvider) FetchDaily(ctx context.Context) (entity.DailyChallenge, error) {
	csrfToken, err := p.initSession(ctx)
	if err != nil {
		log.Printf("⚠ failed to init leetcode session: %v (continuing anyway)", err)
	}

	var lastErr error
	for attempt := range maxRetries {
		if attempt > 0 {
			wait := retryBaseWait * time.Duration(1<<(attempt-1)) // exponential backoff
			log.Printf("⏳ retry %d/%d after %v", attempt+1, maxRetries, wait)
			select {
			case <-ctx.Done():
				return entity.DailyChallenge{}, ctx.Err()
			case <-time.After(wait):
			}
		}

		challenge, err := p.doFetchDaily(ctx, csrfToken)
		if err == nil {
			return challenge, nil
		}
		lastErr = err
		log.Printf("⚠ attempt %d failed: %v", attempt+1, lastErr)
	}

	return entity.DailyChallenge{}, fmt.Errorf("fetch daily challenge: %w (after %d attempts)", lastErr, maxRetries)
}

func (p *LeetCodeProvider) doFetchDaily(ctx context.Context, csrfToken string) (entity.DailyChallenge, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, leetcodeGraphQL, strings.NewReader(dailyQuery))
	if err != nil {
		return entity.DailyChallenge{}, fmt.Errorf("create leetcode request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://leetcode.com")
	req.Header.Set("Origin", "https://leetcode.com")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	if csrfToken != "" {
		req.Header.Set("X-Csrftoken", csrfToken)
	}

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
