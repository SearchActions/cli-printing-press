package types

type ArchivePayload struct {
	Success bool `json:"success"`
}

type Comment struct {
	Id string `json:"id"`
	Body string `json:"body"`
	User string `json:"user"`
	CreatedAt string `json:"createdAt"`
}

type CommentConnection struct {
	Nodes string `json:"nodes"`
}

type CommentPayload struct {
	Success bool `json:"success"`
	Comment string `json:"comment"`
}

type Cycle struct {
	Id string `json:"id"`
	Number int `json:"number"`
	Name string `json:"name"`
	StartsAt string `json:"startsAt"`
	EndsAt string `json:"endsAt"`
	Progress float64 `json:"progress"`
}

type CycleConnection struct {
	Nodes string `json:"nodes"`
}

type DeletePayload struct {
	Success bool `json:"success"`
}

type Document struct {
	Id string `json:"id"`
	Title string `json:"title"`
	Content string `json:"content"`
}

type DocumentConnection struct {
	Nodes string `json:"nodes"`
}

type DocumentPayload struct {
	Success bool `json:"success"`
}

type Issue struct {
	Id string `json:"id"`
	Identifier string `json:"identifier"`
	Title string `json:"title"`
	Description string `json:"description"`
	State string `json:"state"`
	Assignee string `json:"assignee"`
	Priority int `json:"priority"`
	PriorityLabel string `json:"priorityLabel"`
	Project string `json:"project"`
	Cycle string `json:"cycle"`
	Estimate int `json:"estimate"`
	DueDate string `json:"dueDate"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	Url string `json:"url"`
}

type IssueConnection struct {
	Nodes string `json:"nodes"`
	PageInfo string `json:"pageInfo"`
}

type IssuePayload struct {
	Success bool `json:"success"`
	Issue string `json:"issue"`
}

type LabelConnection struct {
	Nodes string `json:"nodes"`
}

type LabelPayload struct {
	Success bool `json:"success"`
}

type NotificationConnection struct {
	Nodes string `json:"nodes"`
}

type Organization struct {
	Id string `json:"id"`
	Name string `json:"name"`
	UrlKey string `json:"urlKey"`
}

type Project struct {
	Id string `json:"id"`
	Name string `json:"name"`
	Description string `json:"description"`
	State string `json:"state"`
	Progress float64 `json:"progress"`
	StartDate string `json:"startDate"`
	TargetDate string `json:"targetDate"`
}

type ProjectConnection struct {
	Nodes string `json:"nodes"`
}

type ProjectPayload struct {
	Success bool `json:"success"`
	Project string `json:"project"`
}

type Team struct {
	Id string `json:"id"`
	Name string `json:"name"`
	Key string `json:"key"`
	Description string `json:"description"`
}

type TeamConnection struct {
	Nodes string `json:"nodes"`
}

type User struct {
	Id string `json:"id"`
	Name string `json:"name"`
	Email string `json:"email"`
	DisplayName string `json:"displayName"`
	Active bool `json:"active"`
}

type UserConnection struct {
	Nodes string `json:"nodes"`
}

type WebhookConnection struct {
	Nodes string `json:"nodes"`
}

type WebhookPayload struct {
	Success bool `json:"success"`
}

type WorkflowStateConnection struct {
	Nodes string `json:"nodes"`
}

