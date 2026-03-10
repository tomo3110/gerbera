package main

const snsCSS = `
/* Page wrapper custom elements (force block display) */
page-home, page-profile, page-post,
page-messages, page-settings, page-search {
	display: block;
}

/* === Reset & Base === */
body {
	background: var(--g-bg);
	color: var(--g-text);
	font-family: var(--g-font);
	margin: 0;
	padding: 0;
}

/* === Layout === */
.sns-layout {
	display: flex;
	min-height: 100vh;
}
.sns-sidebar {
	display: none;
	width: 240px;
	min-height: 100vh;
	border-right: 1px solid var(--g-border);
	background: var(--g-bg-surface);
	padding: var(--g-space-md) 0;
	flex-shrink: 0;
	position: sticky;
	top: 0;
	align-self: flex-start;
}
.sns-main {
	flex: 1;
	min-width: 0;
	max-width: 640px;
	margin: 0 auto;
	padding: var(--g-space-sm);
}

@media (min-width: 769px) {
	.sns-sidebar {
		display: block;
	}
	.sns-main {
		padding: var(--g-space-md);
	}
	.sns-mobile-header {
		display: none !important;
	}
}

/* === Mobile Header === */
.sns-mobile-header {
	display: flex;
	align-items: center;
	gap: var(--g-space-sm);
	padding: var(--g-space-sm) var(--g-space-md);
	border-bottom: 1px solid var(--g-border);
	background: var(--g-bg-surface);
	position: sticky;
	top: 0;
	z-index: 50;
}
.sns-mobile-header h1 {
	margin: 0;
	font-size: 1.2rem;
	flex: 1;
}

/* === Nav Links === */
.sns-nav-link {
	display: flex;
	align-items: center;
	gap: var(--g-space-sm);
	padding: var(--g-space-sm) var(--g-space-md);
	color: var(--g-text);
	text-decoration: none;
	cursor: pointer;
	border: none;
	background: none;
	width: 100%;
	font-size: 1rem;
	font-family: var(--g-font);
	transition: background 0.15s;
}
.sns-nav-link:hover {
	background: var(--g-bg-inset);
}
.sns-nav-link.active {
	color: var(--g-accent);
	font-weight: 600;
}
.sns-nav-badge {
	background: var(--g-danger);
	color: #fff;
	font-size: 0.75rem;
	padding: 1px 6px;
	border-radius: 10px;
	margin-left: auto;
}

/* === Compose Box === */
.compose-box {
	padding: var(--g-space-md);
}
.compose-box textarea {
	resize: none;
	min-height: 80px;
}
.compose-footer {
	display: flex;
	align-items: center;
	justify-content: space-between;
	margin-top: var(--g-space-sm);
}
.compose-footer .char-count {
	font-size: 0.85rem;
	color: var(--g-text-tertiary);
}
.compose-footer .char-count.over {
	color: var(--g-danger);
	font-weight: 600;
}
.compose-actions {
	display: flex;
	align-items: center;
	gap: var(--g-space-sm);
}

/* === Post Card === */
.post-card {
	border-bottom: 1px solid var(--g-border);
	padding: var(--g-space-md);
	transition: background 0.15s;
}
.post-card:hover {
	background: var(--g-bg-inset);
}
.post-header {
	display: flex;
	align-items: center;
	gap: var(--g-space-sm);
	margin-bottom: var(--g-space-xs);
}
.post-author {
	font-weight: 600;
	cursor: pointer;
}
.post-author:hover {
	text-decoration: underline;
}
.post-username {
	color: var(--g-text-tertiary);
	font-size: 0.9rem;
}
.post-time {
	color: var(--g-text-tertiary);
	font-size: 0.85rem;
	margin-left: auto;
}
.post-content {
	display: block;
	margin: var(--g-space-xs) 0;
	line-height: 1.5;
	white-space: pre-wrap;
	word-wrap: break-word;
	color: inherit;
	text-decoration: none;
}
a.post-author {
	color: inherit;
	text-decoration: none;
}
a.post-action-btn {
	text-decoration: none;
	color: var(--g-text-tertiary);
}
a.search-user-item {
	text-decoration: none;
	color: inherit;
}
.post-image {
	width: 100%;
	max-height: 400px;
	object-fit: cover;
	border-radius: var(--g-radius-md);
	margin: var(--g-space-sm) 0;
}
.post-actions {
	display: flex;
	gap: var(--g-space-lg);
	margin-top: var(--g-space-sm);
}
.post-action-btn {
	display: inline-flex;
	align-items: center;
	gap: 4px;
	background: none;
	border: none;
	color: var(--g-text-tertiary);
	cursor: pointer;
	font-size: 0.9rem;
	padding: 4px 8px;
	border-radius: var(--g-radius-sm);
	font-family: var(--g-font);
	transition: color 0.15s, background 0.15s;
}
.post-action-btn:hover {
	background: var(--g-bg-inset);
}
.post-action-btn.liked {
	color: #ef4444;
}
.post-action-btn.retweeted {
	color: #22c55e;
}

/* === Profile === */
.profile-header {
	padding: var(--g-space-md);
	border-bottom: 1px solid var(--g-border);
}
.profile-info {
	display: flex;
	align-items: flex-start;
	gap: var(--g-space-md);
	margin-bottom: var(--g-space-md);
}
.profile-names {
	flex: 1;
}
.profile-names h2 {
	margin: 0 0 2px;
}
.profile-names .username {
	color: var(--g-text-tertiary);
	font-size: 0.9rem;
}
.profile-bio {
	margin: var(--g-space-sm) 0;
	color: var(--g-text-secondary);
}
.profile-stats {
	display: flex;
	gap: var(--g-space-lg);
}
.profile-stat {
	text-align: center;
}
.profile-stat strong {
	display: block;
	font-size: 1.1rem;
}
.profile-stat span {
	color: var(--g-text-tertiary);
	font-size: 0.85rem;
}

/* === Comment === */
.comment-item {
	display: flex;
	gap: var(--g-space-sm);
	padding: var(--g-space-sm) var(--g-space-md);
	border-bottom: 1px solid var(--g-border);
}
.comment-body {
	flex: 1;
}
.comment-author {
	font-weight: 600;
	font-size: 0.9rem;
}
.comment-text {
	margin: 2px 0;
	line-height: 1.4;
}
.comment-time {
	color: var(--g-text-tertiary);
	font-size: 0.8rem;
}
.comment-form {
	padding: var(--g-space-md);
	display: flex;
	gap: var(--g-space-sm);
	align-items: flex-end;
}
.comment-form textarea {
	resize: none;
	min-height: 40px;
	flex: 1;
}

/* === Messages === */
.convo-item {
	display: flex;
	align-items: center;
	gap: var(--g-space-sm);
	padding: var(--g-space-sm) var(--g-space-md);
	border-bottom: 1px solid var(--g-border);
	cursor: pointer;
	transition: background 0.15s;
}
.convo-item:hover {
	background: var(--g-bg-inset);
}
.convo-info {
	flex: 1;
	min-width: 0;
}
.convo-name {
	font-weight: 600;
	font-size: 0.95rem;
}
.convo-preview {
	color: var(--g-text-tertiary);
	font-size: 0.85rem;
	white-space: nowrap;
	overflow: hidden;
	text-overflow: ellipsis;
}
.convo-meta {
	text-align: right;
	flex-shrink: 0;
}
.convo-time {
	color: var(--g-text-tertiary);
	font-size: 0.8rem;
}
.chat-back-btn {
	background: none;
	border: none;
	cursor: pointer;
	color: var(--g-accent);
	font-size: 0.9rem;
	padding: var(--g-space-xs) var(--g-space-sm);
	font-family: var(--g-font);
}

/* === Search === */
.search-input-wrap {
	padding: var(--g-space-md);
}
.search-section-label {
	padding: var(--g-space-xs) var(--g-space-md);
	font-size: 0.85rem;
	font-weight: 600;
	color: var(--g-text-tertiary);
	text-transform: uppercase;
	letter-spacing: 0.05em;
	border-bottom: 1px solid var(--g-border);
}
.search-user-item {
	display: flex;
	align-items: center;
	gap: var(--g-space-sm);
	padding: var(--g-space-sm) var(--g-space-md);
	border-bottom: 1px solid var(--g-border);
}
.search-user-info {
	flex: 1;
}
.search-user-info strong {
	display: block;
}
.search-user-info span {
	color: var(--g-text-tertiary);
	font-size: 0.85rem;
}

/* === Settings === */
.settings-section {
	padding: var(--g-space-md);
	border-bottom: 1px solid var(--g-border);
}
.settings-section h3 {
	margin: 0 0 var(--g-space-md);
	font-size: 1rem;
}

/* === SVG Icons (inline) === */
.post-action-btn > span:first-child {
	display: inline-flex;
	align-items: center;
}
.icon-heart, .icon-retweet, .icon-comment, .icon-share {
	width: 18px;
	height: 18px;
	display: inline-block;
	vertical-align: middle;
}

/* === Utility === */
.text-center { text-align: center; }
.text-secondary { color: var(--g-text-secondary); }
.text-tertiary { color: var(--g-text-tertiary); }
.mt-sm { margin-top: var(--g-space-sm); }
.mb-sm { margin-bottom: var(--g-space-sm); }
.full-width { width: 100%; }
`
