package main

import (
	"time"

	"github.com/Le-Chat99/blogAggregator/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	ApiKey    string    `json:"api_key"`
}

func databaseUserToUser(user database.User) User {
	return User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Name:      user.Name,
		ApiKey:    user.ApiKey,
	}
}

type Feed struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	Url       string    `json:"url"`
	UserID    uuid.UUID `json:"user_id"`
}

func databaseFeedToFeed(feed database.Feed) Feed {
	return Feed{
		ID:        feed.ID,
		CreatedAt: feed.CreatedAt,
		UpdatedAt: feed.UpdatedAt,
		Name:      feed.Name,
		Url:       feed.Url,
		UserID:    feed.UserID,
	}
}

type FeedFollowed struct {
	ID        uuid.UUID `json:"id"`
	FeedID    uuid.UUID `json:"feed_id"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func databaseFeedFollowedToFeedFollowed(feed database.FeedFollow) FeedFollowed {
	return FeedFollowed{
		ID:        feed.ID,
		FeedID:    feed.FeedID,
		UserID:    feed.UserID,
		CreatedAt: feed.CreatedAt,
		UpdatedAt: feed.UpdatedAt,
	}
}

func databaseFeedFollowsToFeedFollows(feedFollows []database.FeedFollow) []FeedFollowed {
	result := make([]FeedFollowed, len(feedFollows))
	for i, feedFollow := range feedFollows {
		result[i] = databaseFeedFollowedToFeedFollowed(feedFollow)
	}
	return result
}

type FeedData struct {
	Channel struct {
		Title       string
		Description string
		Items       []struct {
			Title       string
			Description string
		} `xml:"item"`
	} `xml:"channel"`
}
