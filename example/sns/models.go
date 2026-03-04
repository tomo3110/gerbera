package main

import "time"

type User struct {
	ID          int64
	Username    string
	DisplayName string
	Email       string
	PasswordHash string
	Bio         string
	AvatarPath  string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Post struct {
	ID        int64
	UserID    int64
	Content   string
	ImagePath string
	CreatedAt time.Time
}

type PostWithMeta struct {
	Post
	Author       User
	LikeCount    int
	RetweetCount int
	CommentCount int
	Liked        bool
	Retweeted    bool
}

type Like struct {
	ID        int64
	UserID    int64
	PostID    int64
	CreatedAt time.Time
}

type Retweet struct {
	ID        int64
	UserID    int64
	PostID    int64
	CreatedAt time.Time
}

type Follow struct {
	ID          int64
	FollowerID  int64
	FollowingID int64
	CreatedAt   time.Time
}

type Comment struct {
	ID        int64
	UserID    int64
	PostID    int64
	Content   string
	CreatedAt time.Time
}

type CommentWithUser struct {
	Comment
	Author User
}

type Message struct {
	ID         int64
	SenderID   int64
	ReceiverID int64
	Content    string
	ReadAt     *time.Time
	CreatedAt  time.Time
}

type ConversationSummary struct {
	Partner       User
	LastMessage   string
	LastMessageAt time.Time
	UnreadCount   int
}

type Notification struct {
	ID        int64
	UserID    int64
	ActorID   int64
	Type      string
	PostID    *int64
	ReadAt    *time.Time
	CreatedAt time.Time
}
