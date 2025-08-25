package jira

const (
	TimeFormat = "2006-01-02T15:04:05Z0700"
)

type Issue struct {
	// 이슈 ID. e.g., "185730"
	ID string `json:"id"`
	// 이슈 번호.
	Key string `json:"key"`
	// 이슈 필드 정보.
	Fields IssueFields `json:"fields"`
}

type IssueFields struct {
	// 이슈 제목.
	Title string `json:"summary"`
	// 이슈 본문 내용.
	Content string `json:"description"`
	// 레이블 목록.
	Labels []string `json:"labels"`
	// 이슈 생성자.
	Creator User `json:"creator"`
	// 이슈 할당자.
	Assignee User `json:"assignee"`
	// 이슈 상태.
	Status Status `json:"status"`
	// 코멘트 정보.
	CommentInfo CommentInfo `json:"comment"`
	// 이슈 생성일.
	Created string `json:"created"`
	// 이슈 수정일.
	Updated string `json:"updated"`
}

type User struct {
	// 사용자 ID.
	ID string `json:"name"`
	// 사용자 표시 이름.
	Name string `json:"displayName"`
	// 이메일 주소.
	EmailAddress string `json:"emailAddress"`
}

type Status struct {
	// 상태 이름. e.g., "In Progress", "Resolved"
	Name string `json:"name"`
}

type CommentInfo struct {
	// 코멘트 총 개수.
	Total int `json:"total"`
	// 코멘트 목록.
	Comments []Comment `json:"comments"`
}

type Comment struct {
	// 코멘트 ID. e.g., "1044263"
	ID string `json:"id"`
	// 코멘트 작성자.
	Author User `json:"author"`
	// 코멘트 내용.
	Body string `json:"body"`
	// 코멘트 작성일. (2006-01-02T15:04:05Z0700)
	Created string `json:"created"`
	// 코멘트 수정일. (2006-01-02T15:04:05Z0700)
	Updated string `json:"updated"`
}
