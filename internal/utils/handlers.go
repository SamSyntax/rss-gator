package utils

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/SamSyntax/Gator/internal/config"
	"github.com/SamSyntax/Gator/internal/database"
	"github.com/google/uuid"
)

type State struct {
	Db  *database.Queries
	Cfg *config.Config
}

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	Names map[string]func(*State, Command) error
}

// Commands

func (c *Commands) Register(name string, f func(*State, Command) error) {
	if c.Names == nil {
		c.Names = make(map[string]func(*State, Command) error)
	}
	c.Names[name] = f
}

func (c *Commands) Run(s *State, cmd Command) error {
	handler, exists := c.Names[cmd.Name]
	if !exists {
		return fmt.Errorf("Unknown command: %s", cmd.Name)
	}

	return handler(s, cmd)
}

// Proper handlers

func HandlerRegister(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return errors.New("Args slice is empty")
	}
	if s.Cfg.CURRENT_USER_NAME == "" {
		return errors.New("You must set current user name in config")
	}
	username := cmd.Args[0]
	user, err := s.Db.GetUser(context.Background(), username)
	if err != nil {
		log.Printf("Failed to fetch user from database: %v", err)
	}
	if user.Name == username {
		return errors.New("This user already exists in the database: %v")
	}
	_, err = s.Db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		Name:      username,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		return errors.Join(errors.New("Failed to create user"), err)
	}
	err = config.SetUser(username)
	if err != nil {
		return errors.Join(errors.New("Failed to write a config change"), err)
	}
	if err != nil {
		return fmt.Errorf("Failed to create user: %v", err)
	}
	fmt.Println("User has been created")
	return nil
}

func HandlerUsers(s *State, cmd Command) error {
	users, err := s.Db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("Failed to fetch users from database: %v", err)
	}

	for _, user := range users {
		if user.Name == s.Cfg.CURRENT_USER_NAME {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}

	return nil
}

func HandlerDelete(s *State, cmd Command) error {
	err := s.Db.DeleteExec(context.Background())
	if err != nil {
		return fmt.Errorf("Failed to delete users: %v", err)
	}

	fmt.Println("Users have been deleted")
	return nil
}

func HandlerLogin(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return errors.New("Args slice is empty")
	}
	if s.Cfg.CURRENT_USER_NAME == "" {
		return errors.New("You must set current user name in config")
	}
	username := cmd.Args[0]
	user, err := s.Db.GetUser(context.Background(), username)
	if err != nil {
		return fmt.Errorf("Failed to get user from database: %v", err)
	}
	err = config.SetUser(user.Name)
	if err != nil {
		return fmt.Errorf("Failed to set user: %v", err)
	}
	cmd.Name = s.Cfg.CURRENT_USER_NAME

	fmt.Printf("User %s has been set", user.Name)

	return nil
}

func HandlerAgg(s *State, cmd Command) error {
	StartScrape(s.Db, 15, time.Minute)
	return nil
}

func HandlerAddFeed(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) == 0 {
		return errors.New("Args slice is empty")
	}
	if len(cmd.Args) < 2 {
		return errors.New("This command needs 2 arguments [name_of_the_feed] [url_to_the_feed]")
	}
	ctx := context.Background()

	feed, err := s.Db.CreateFeed(ctx, database.CreateFeedParams{
		ID:        uuid.New(),
		UserID:    user.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.Args[0],
		Url:       cmd.Args[1],
	})
	if err != nil {
		return errors.Join(fmt.Errorf("There was an error creating feed: "), err)
	}
	_, err = s.Db.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return errors.Join(fmt.Errorf("There was an error creating feed follow: "), err)
	}
	fmt.Println(feed.ID, feed.UserID)
	return nil
}

func HandlerGetFeeds(s *State, cmd Command) error {
	feeds, err := s.Db.GetFeeds(context.Background())
	if err != nil {
		return errors.Join(fmt.Errorf("There was an error fetching feeds: "), err)
	}

	for _, feed := range feeds {
		user, err := s.Db.GetUserById(context.Background(), feed.UserID)
		if err != nil {
			return errors.Join(fmt.Errorf("There was an error fetching user: "), err)
		}
		fmt.Println(feed.Name, feed.Url, feed.UserID, user.Name)
	}
	return nil
}

// Feed Follows

func HandlerFollowFeed(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) != 1 {
		return errors.New("This command accepts only one argument [feed_url]")
	}
	ctx := context.Background()

	feed, err := s.Db.GetFeedByUrl(ctx, cmd.Args[0])
	if err != nil {
		return errors.Join(fmt.Errorf("Coudln't find feed in database: "), err)
	}
	if err != nil {
		return errors.Join(fmt.Errorf("Coudln't find user in database: "), err)
	}
	follow, err := s.Db.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return errors.Join(fmt.Errorf("Coudln't create feed follow: "), err)
	}

	fmt.Printf("Followed feed %s | follow ID %s\n", feed.Name, follow.ID)

	return nil
}

func HandlerGetFollowedFeeds(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) > 0 {
		return errors.New("This command does not accept any arguments")
	}
	ctx := context.Background()
	follows, err := s.Db.GetFeedFollows(ctx, user.ID)
	if err != nil {
		return errors.Join(fmt.Errorf("Coudln't fetch follows from database: "), err)
	}
	if len(follows) == 0 {
		fmt.Printf("User %s does not follow any feeds", user.Name)
	}

	for _, follow := range follows {
		feed, err := s.Db.GetFeedById(ctx, follow.FeedID)
		if err != nil {
			fmt.Printf("Coudln't fetch follows from database: %v", err)
			continue
		}
		fmt.Println(feed.Name)
	}

	return nil
}

func HandlerDeleteFollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) != 1 {
		return errors.New("This command accepts only one argument [feed_url]")
	}
	ctx := context.Background()
	feed, err := s.Db.GetFeedByUrl(ctx, cmd.Args[0])
	if err != nil {
		return errors.Join(errors.New("Couldn't find feed in database"), err)
	}
	err = s.Db.DeleteFeedFollow(ctx, database.DeleteFeedFollowParams{
		FeedID: feed.ID,
		UserID: user.ID,
	})
	if err != nil {
		return errors.Join(errors.New("Couldn't unfollow feed"), err)
	}
	return nil
}
