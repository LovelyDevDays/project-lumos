package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	maxResults = 100
	timeout    = 300 * time.Second
	maxRetries = 3
	outputDir  = "output"
	outputFile = "gs_issues.json"
)

type JiraResponse struct {
	Issues []map[string]any `json:"issues"`
	Total  int              `json:"total"`
}

type Config struct {
	Token   string
	BaseURL string
	Project string
}

func main() {
	// 설정 로드
	config, err := loadConfig()
	if err != nil {
		fmt.Printf("❌ 설정 로드 실패: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("JIRA_BASE_URL:", config.BaseURL)
	fmt.Println("JIRA_PROJECT:", config.Project)

	fmt.Printf("🚀 JIRA 이슈 수집 시작... (프로젝트: %s)\n", config.Project)

	client := &http.Client{
		Timeout: timeout,
	}

	var allIssues []map[string]any
	startAt := 0

	for {
		issues, err := fetchIssues(client, config, startAt)
		if err != nil {
			fmt.Printf("❌ 이슈 수집 실패: %v\n", err)
			os.Exit(1)
		}

		allIssues = append(allIssues, issues...)
		fmt.Printf("📥 %d ~ %d번까지 수집 완료\n", startAt, startAt+len(issues))

		if len(issues) < maxResults {
			break
		}

		startAt += maxResults
	}

	// 결과 저장
	if err := saveIssues(allIssues); err != nil {
		fmt.Printf("❌ 파일 저장 실패: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✅ 총 %d개의 이슈를 다음 경로에 저장했습니다:\n%s\n",
		len(allIssues), filepath.Join(outputDir, outputFile))
}

// 설정 로드
func loadConfig() (*Config, error) {
	token := os.Getenv("JIRA_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("JIRA_TOKEN 환경변수가 설정되지 않았습니다")
	}

	baseURL := os.Getenv("JIRA_BASE_URL")
	if baseURL == "" {
		return nil, fmt.Errorf("JIRA_BASE_URL 환경변수가 설정되지 않았습니다")
	}

	project := os.Getenv("JIRA_PROJECT")
	if project == "" {
		return nil, fmt.Errorf("JIRA_PROJECT 환경변수가 설정되지 않았습니다")
	}

	return &Config{
		Token:   token,
		BaseURL: baseURL,
		Project: project,
	}, nil
}

func fetchIssues(client *http.Client, config *Config, startAt int) ([]map[string]any, error) {
	var issues []map[string]any

	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := createRequest(config, startAt)
		if err != nil {
			return nil, fmt.Errorf("요청 생성 실패: %w", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("⚠️  %d~ 요청 실패 (시도 %d/%d): %v\n", startAt, attempt, maxRetries, err)
			if attempt == maxRetries {
				return nil, fmt.Errorf("최대 재시도 횟수 초과")
			}
			time.Sleep(3 * time.Second)
			continue
		}

		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("⚠️  HTTP 오류 (시도 %d/%d): %d %s\n", attempt, maxRetries, resp.StatusCode, resp.Status)
			if attempt == maxRetries {
				return nil, fmt.Errorf("HTTP 오류: %d %s", resp.StatusCode, resp.Status)
			}
			time.Sleep(3 * time.Second)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("⚠️  응답 읽기 실패 (시도 %d/%d): %v\n", attempt, maxRetries, err)
			if attempt == maxRetries {
				return nil, fmt.Errorf("응답 읽기 실패: %w", err)
			}
			time.Sleep(3 * time.Second)
			continue
		}

		var jiraResp JiraResponse
		if err := json.Unmarshal(body, &jiraResp); err != nil {
			fmt.Printf("⚠️  JSON 파싱 실패 (시도 %d/%d): %v\n", attempt, maxRetries, err)
			if attempt == maxRetries {
				return nil, fmt.Errorf("JSON 파싱 실패: %w", err)
			}
			time.Sleep(3 * time.Second)
			continue
		}

		return jiraResp.Issues, nil
	}

	return issues, nil
}

func createRequest(config *Config, startAt int) (*http.Request, error) {
	u, err := url.Parse(config.BaseURL)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("jql", fmt.Sprintf("project=%s", config.Project))
	q.Set("startAt", strconv.Itoa(startAt))
	q.Set("maxResults", strconv.Itoa(maxResults))
	q.Set("fields", "*all")
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+config.Token)
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func saveIssues(issues []map[string]any) error {
	// 출력 디렉토리 생성
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("디렉토리 생성 실패: %w", err)
	}

	// JSON 파일로 저장
	outputPath := filepath.Join(outputDir, outputFile)
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("파일 생성 실패: %w", err)
	}
	defer func() { _ = file.Close() }()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(issues); err != nil {
		return fmt.Errorf("JSON 인코딩 실패: %w", err)
	}

	return nil
}
