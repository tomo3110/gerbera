package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Retry connection up to 30 times (for Docker Compose startup ordering)
	for i := 0; i < 30; i++ {
		if err := db.Ping(); err == nil {
			return db, nil
		}
		log.Printf("waiting for MySQL... (%d/30)", i+1)
		time.Sleep(2 * time.Second)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("MySQL not ready after 60s: %w", err)
	}
	return db, nil
}

func autoMigrate(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(30) NOT NULL UNIQUE,
			display_name VARCHAR(50) NOT NULL,
			email VARCHAR(255) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			bio VARCHAR(160) DEFAULT '',
			avatar_path VARCHAR(255) DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS posts (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT NOT NULL,
			content VARCHAR(250) NOT NULL,
			image_path VARCHAR(255) DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS likes (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT NOT NULL,
			post_id BIGINT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY uk_likes (user_id, post_id),
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (post_id) REFERENCES posts(id)
		)`,
		`CREATE TABLE IF NOT EXISTS retweets (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT NOT NULL,
			post_id BIGINT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY uk_retweets (user_id, post_id),
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (post_id) REFERENCES posts(id)
		)`,
		`CREATE TABLE IF NOT EXISTS follows (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			follower_id BIGINT NOT NULL,
			following_id BIGINT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY uk_follows (follower_id, following_id),
			FOREIGN KEY (follower_id) REFERENCES users(id),
			FOREIGN KEY (following_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS comments (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT NOT NULL,
			post_id BIGINT NOT NULL,
			content VARCHAR(250) NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (post_id) REFERENCES posts(id)
		)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			sender_id BIGINT NOT NULL,
			receiver_id BIGINT NOT NULL,
			content VARCHAR(500) NOT NULL,
			read_at DATETIME DEFAULT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (sender_id) REFERENCES users(id),
			FOREIGN KEY (receiver_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS notifications (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT NOT NULL,
			actor_id BIGINT NOT NULL,
			type VARCHAR(20) NOT NULL,
			post_id BIGINT DEFAULT NULL,
			read_at DATETIME DEFAULT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (actor_id) REFERENCES users(id)
		)`,
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return fmt.Errorf("migration: %w", err)
		}
	}
	return nil
}

// --- User queries ---

func dbCreateUser(db *sql.DB, username, displayName, email, passwordHash string) (int64, error) {
	res, err := db.Exec(
		`INSERT INTO users (username, display_name, email, password_hash) VALUES (?, ?, ?, ?)`,
		username, displayName, email, passwordHash,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func dbGetUserByUsername(db *sql.DB, username string) (*User, error) {
	u := &User{}
	err := db.QueryRow(
		`SELECT id, username, display_name, email, password_hash, bio, avatar_path, created_at, updated_at FROM users WHERE username = ?`,
		username,
	).Scan(&u.ID, &u.Username, &u.DisplayName, &u.Email, &u.PasswordHash, &u.Bio, &u.AvatarPath, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func dbGetUserByID(db *sql.DB, id int64) (*User, error) {
	u := &User{}
	err := db.QueryRow(
		`SELECT id, username, display_name, email, password_hash, bio, avatar_path, created_at, updated_at FROM users WHERE id = ?`,
		id,
	).Scan(&u.ID, &u.Username, &u.DisplayName, &u.Email, &u.PasswordHash, &u.Bio, &u.AvatarPath, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func dbGetUserByEmail(db *sql.DB, email string) (*User, error) {
	u := &User{}
	err := db.QueryRow(
		`SELECT id, username, display_name, email, password_hash, bio, avatar_path, created_at, updated_at FROM users WHERE email = ?`,
		email,
	).Scan(&u.ID, &u.Username, &u.DisplayName, &u.Email, &u.PasswordHash, &u.Bio, &u.AvatarPath, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func dbUpdateUserProfile(db *sql.DB, id int64, displayName, email, bio string) error {
	_, err := db.Exec(
		`UPDATE users SET display_name = ?, email = ?, bio = ? WHERE id = ?`,
		displayName, email, bio, id,
	)
	return err
}

func dbUpdateUserPassword(db *sql.DB, id int64, passwordHash string) error {
	_, err := db.Exec(`UPDATE users SET password_hash = ? WHERE id = ?`, passwordHash, id)
	return err
}

func dbUpdateUserAvatar(db *sql.DB, id int64, avatarPath string) error {
	_, err := db.Exec(`UPDATE users SET avatar_path = ? WHERE id = ?`, avatarPath, id)
	return err
}

func dbSearchUsers(db *sql.DB, query string, limit int) ([]User, error) {
	rows, err := db.Query(
		`SELECT id, username, display_name, email, password_hash, bio, avatar_path, created_at, updated_at
		 FROM users WHERE username LIKE ? OR display_name LIKE ? LIMIT ?`,
		"%"+query+"%", "%"+query+"%", limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.DisplayName, &u.Email, &u.PasswordHash, &u.Bio, &u.AvatarPath, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// --- Post queries ---

func dbCreatePost(db *sql.DB, userID int64, content, imagePath string) (int64, error) {
	res, err := db.Exec(
		`INSERT INTO posts (user_id, content, image_path) VALUES (?, ?, ?)`,
		userID, content, imagePath,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func dbDeletePost(db *sql.DB, postID, userID int64) error {
	_, err := db.Exec(`DELETE FROM posts WHERE id = ? AND user_id = ?`, postID, userID)
	return err
}

func dbGetTimeline(db *sql.DB, userID int64, limit, offset int) ([]PostWithMeta, error) {
	rows, err := db.Query(`
		SELECT p.id, p.user_id, p.content, p.image_path, p.created_at,
			u.id, u.username, u.display_name, u.avatar_path,
			(SELECT COUNT(*) FROM likes WHERE post_id = p.id) AS like_count,
			(SELECT COUNT(*) FROM retweets WHERE post_id = p.id) AS retweet_count,
			(SELECT COUNT(*) FROM comments WHERE post_id = p.id) AS comment_count,
			EXISTS(SELECT 1 FROM likes WHERE post_id = p.id AND user_id = ?) AS liked,
			EXISTS(SELECT 1 FROM retweets WHERE post_id = p.id AND user_id = ?) AS retweeted
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.user_id = ? OR p.user_id IN (SELECT following_id FROM follows WHERE follower_id = ?)
		ORDER BY p.created_at DESC
		LIMIT ? OFFSET ?`,
		userID, userID, userID, userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPostsWithMeta(rows)
}

func dbGetUserPosts(db *sql.DB, profileUserID, viewerID int64, limit, offset int) ([]PostWithMeta, error) {
	rows, err := db.Query(`
		SELECT p.id, p.user_id, p.content, p.image_path, p.created_at,
			u.id, u.username, u.display_name, u.avatar_path,
			(SELECT COUNT(*) FROM likes WHERE post_id = p.id) AS like_count,
			(SELECT COUNT(*) FROM retweets WHERE post_id = p.id) AS retweet_count,
			(SELECT COUNT(*) FROM comments WHERE post_id = p.id) AS comment_count,
			EXISTS(SELECT 1 FROM likes WHERE post_id = p.id AND user_id = ?) AS liked,
			EXISTS(SELECT 1 FROM retweets WHERE post_id = p.id AND user_id = ?) AS retweeted
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.user_id = ?
		ORDER BY p.created_at DESC
		LIMIT ? OFFSET ?`,
		viewerID, viewerID, profileUserID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPostsWithMeta(rows)
}

func dbGetPost(db *sql.DB, postID, viewerID int64) (*PostWithMeta, error) {
	row := db.QueryRow(`
		SELECT p.id, p.user_id, p.content, p.image_path, p.created_at,
			u.id, u.username, u.display_name, u.avatar_path,
			(SELECT COUNT(*) FROM likes WHERE post_id = p.id) AS like_count,
			(SELECT COUNT(*) FROM retweets WHERE post_id = p.id) AS retweet_count,
			(SELECT COUNT(*) FROM comments WHERE post_id = p.id) AS comment_count,
			EXISTS(SELECT 1 FROM likes WHERE post_id = p.id AND user_id = ?) AS liked,
			EXISTS(SELECT 1 FROM retweets WHERE post_id = p.id AND user_id = ?) AS retweeted
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.id = ?`,
		viewerID, viewerID, postID,
	)
	pm := &PostWithMeta{}
	err := row.Scan(
		&pm.Post.ID, &pm.Post.UserID, &pm.Post.Content, &pm.Post.ImagePath, &pm.Post.CreatedAt,
		&pm.Author.ID, &pm.Author.Username, &pm.Author.DisplayName, &pm.Author.AvatarPath,
		&pm.LikeCount, &pm.RetweetCount, &pm.CommentCount, &pm.Liked, &pm.Retweeted,
	)
	if err != nil {
		return nil, err
	}
	return pm, nil
}

func dbSearchPosts(db *sql.DB, query string, viewerID int64, limit int) ([]PostWithMeta, error) {
	rows, err := db.Query(`
		SELECT p.id, p.user_id, p.content, p.image_path, p.created_at,
			u.id, u.username, u.display_name, u.avatar_path,
			(SELECT COUNT(*) FROM likes WHERE post_id = p.id) AS like_count,
			(SELECT COUNT(*) FROM retweets WHERE post_id = p.id) AS retweet_count,
			(SELECT COUNT(*) FROM comments WHERE post_id = p.id) AS comment_count,
			EXISTS(SELECT 1 FROM likes WHERE post_id = p.id AND user_id = ?) AS liked,
			EXISTS(SELECT 1 FROM retweets WHERE post_id = p.id AND user_id = ?) AS retweeted
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.content LIKE ?
		ORDER BY p.created_at DESC
		LIMIT ?`,
		viewerID, viewerID, "%"+query+"%", limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPostsWithMeta(rows)
}

func scanPostsWithMeta(rows *sql.Rows) ([]PostWithMeta, error) {
	var posts []PostWithMeta
	for rows.Next() {
		var pm PostWithMeta
		if err := rows.Scan(
			&pm.Post.ID, &pm.Post.UserID, &pm.Post.Content, &pm.Post.ImagePath, &pm.Post.CreatedAt,
			&pm.Author.ID, &pm.Author.Username, &pm.Author.DisplayName, &pm.Author.AvatarPath,
			&pm.LikeCount, &pm.RetweetCount, &pm.CommentCount, &pm.Liked, &pm.Retweeted,
		); err != nil {
			return nil, err
		}
		posts = append(posts, pm)
	}
	return posts, rows.Err()
}

// --- Like queries ---

func dbToggleLike(db *sql.DB, userID, postID int64) (liked bool, err error) {
	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM likes WHERE user_id = ? AND post_id = ?`, userID, postID).Scan(&count)
	if err != nil {
		return false, err
	}
	if count > 0 {
		_, err = db.Exec(`DELETE FROM likes WHERE user_id = ? AND post_id = ?`, userID, postID)
		return false, err
	}
	_, err = db.Exec(`INSERT INTO likes (user_id, post_id) VALUES (?, ?)`, userID, postID)
	return true, err
}

// --- Retweet queries ---

func dbToggleRetweet(db *sql.DB, userID, postID int64) (retweeted bool, err error) {
	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM retweets WHERE user_id = ? AND post_id = ?`, userID, postID).Scan(&count)
	if err != nil {
		return false, err
	}
	if count > 0 {
		_, err = db.Exec(`DELETE FROM retweets WHERE user_id = ? AND post_id = ?`, userID, postID)
		return false, err
	}
	_, err = db.Exec(`INSERT INTO retweets (user_id, post_id) VALUES (?, ?)`, userID, postID)
	return true, err
}

// --- Follow queries ---

func dbToggleFollow(db *sql.DB, followerID, followingID int64) (following bool, err error) {
	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM follows WHERE follower_id = ? AND following_id = ?`, followerID, followingID).Scan(&count)
	if err != nil {
		return false, err
	}
	if count > 0 {
		_, err = db.Exec(`DELETE FROM follows WHERE follower_id = ? AND following_id = ?`, followerID, followingID)
		return false, err
	}
	_, err = db.Exec(`INSERT INTO follows (follower_id, following_id) VALUES (?, ?)`, followerID, followingID)
	return true, err
}

func dbIsFollowing(db *sql.DB, followerID, followingID int64) (bool, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM follows WHERE follower_id = ? AND following_id = ?`, followerID, followingID).Scan(&count)
	return count > 0, err
}

func dbFollowerCount(db *sql.DB, userID int64) (int, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM follows WHERE following_id = ?`, userID).Scan(&count)
	return count, err
}

func dbFollowingCount(db *sql.DB, userID int64) (int, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM follows WHERE follower_id = ?`, userID).Scan(&count)
	return count, err
}

// --- Comment queries ---

func dbCreateComment(db *sql.DB, userID, postID int64, content string) (int64, error) {
	res, err := db.Exec(
		`INSERT INTO comments (user_id, post_id, content) VALUES (?, ?, ?)`,
		userID, postID, content,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func dbGetComments(db *sql.DB, postID int64) ([]CommentWithUser, error) {
	rows, err := db.Query(`
		SELECT c.id, c.user_id, c.post_id, c.content, c.created_at,
			u.id, u.username, u.display_name, u.avatar_path
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.post_id = ?
		ORDER BY c.created_at ASC`, postID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var comments []CommentWithUser
	for rows.Next() {
		var cw CommentWithUser
		if err := rows.Scan(
			&cw.Comment.ID, &cw.Comment.UserID, &cw.Comment.PostID, &cw.Comment.Content, &cw.Comment.CreatedAt,
			&cw.Author.ID, &cw.Author.Username, &cw.Author.DisplayName, &cw.Author.AvatarPath,
		); err != nil {
			return nil, err
		}
		comments = append(comments, cw)
	}
	return comments, rows.Err()
}

// --- Message queries ---

func dbSendMessage(db *sql.DB, senderID, receiverID int64, content string) (int64, error) {
	res, err := db.Exec(
		`INSERT INTO messages (sender_id, receiver_id, content) VALUES (?, ?, ?)`,
		senderID, receiverID, content,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func dbGetConversations(db *sql.DB, userID int64) ([]ConversationSummary, error) {
	rows, err := db.Query(`
		SELECT u.id, u.username, u.display_name, u.avatar_path,
			m.content, m.created_at,
			(SELECT COUNT(*) FROM messages WHERE sender_id = u.id AND receiver_id = ? AND read_at IS NULL) AS unread
		FROM (
			SELECT CASE WHEN sender_id = ? THEN receiver_id ELSE sender_id END AS partner_id,
				MAX(id) AS last_msg_id
			FROM messages
			WHERE sender_id = ? OR receiver_id = ?
			GROUP BY partner_id
		) conv
		JOIN users u ON u.id = conv.partner_id
		JOIN messages m ON m.id = conv.last_msg_id
		ORDER BY m.created_at DESC`,
		userID, userID, userID, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var convos []ConversationSummary
	for rows.Next() {
		var cs ConversationSummary
		if err := rows.Scan(
			&cs.Partner.ID, &cs.Partner.Username, &cs.Partner.DisplayName, &cs.Partner.AvatarPath,
			&cs.LastMessage, &cs.LastMessageAt, &cs.UnreadCount,
		); err != nil {
			return nil, err
		}
		convos = append(convos, cs)
	}
	return convos, rows.Err()
}

func dbGetMessages(db *sql.DB, userID, partnerID int64, limit int) ([]Message, error) {
	rows, err := db.Query(`
		SELECT id, sender_id, receiver_id, content, read_at, created_at
		FROM messages
		WHERE (sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)
		ORDER BY created_at ASC
		LIMIT ?`,
		userID, partnerID, partnerID, userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var msgs []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.SenderID, &m.ReceiverID, &m.Content, &m.ReadAt, &m.CreatedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}

func dbMarkMessagesRead(db *sql.DB, receiverID, senderID int64) error {
	_, err := db.Exec(
		`UPDATE messages SET read_at = NOW() WHERE receiver_id = ? AND sender_id = ? AND read_at IS NULL`,
		receiverID, senderID,
	)
	return err
}

// --- Notification queries ---

func dbCreateNotification(db *sql.DB, userID, actorID int64, typ string, postID *int64) error {
	_, err := db.Exec(
		`INSERT INTO notifications (user_id, actor_id, type, post_id) VALUES (?, ?, ?, ?)`,
		userID, actorID, typ, postID,
	)
	return err
}

func dbUnreadNotificationCount(db *sql.DB, userID int64) (int, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM notifications WHERE user_id = ? AND read_at IS NULL`, userID).Scan(&count)
	return count, err
}

func dbUnreadMessageCount(db *sql.DB, userID int64) (int, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM messages WHERE receiver_id = ? AND read_at IS NULL`, userID).Scan(&count)
	return count, err
}

func dbPostCount(db *sql.DB, userID int64) (int, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM posts WHERE user_id = ?`, userID).Scan(&count)
	return count, err
}
