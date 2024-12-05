package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/SamSyntax/Gator/internal/auth"
	"github.com/SamSyntax/Gator/internal/config"
	"github.com/SamSyntax/Gator/internal/database"
	"github.com/SamSyntax/Gator/internal/utils"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.Read()

	db, err := sql.Open("postgres", cfg.DB_URL)
	if err != nil {
		log.Printf("Failed to open DB connection: %v\n", err)
	}
	dbQueries := database.New(db)
	state := &utils.State{Cfg: &cfg, Db: dbQueries}
	commands := &utils.Commands{}
	commands.Register("login", utils.HandlerLogin)
	commands.Register("register", utils.HandlerRegister)
	commands.Register("reset", utils.HandlerDelete)
	commands.Register("users", utils.HandlerUsers)
	commands.Register("agg", utils.HandlerAgg)
	commands.Register("addfeed", auth.MiddlewareLoggedIn(utils.HandlerAddFeed))
	commands.Register("feeds", utils.HandlerGetFeeds)
	commands.Register("follow", auth.MiddlewareLoggedIn(utils.HandlerFollowFeed))
	commands.Register("following", auth.MiddlewareLoggedIn(utils.HandlerGetFollowedFeeds))
	commands.Register("unfollow", auth.MiddlewareLoggedIn(utils.HandlerDeleteFollow))

	args := os.Args
	if len(args) < 2 {
		fmt.Println("Error: not enough arguments provided")
		os.Exit(1)
	}
	cmdName := args[1]
	cmdArgs := args[2:]

	cmd := utils.Command{
		Name: cmdName,
		Args: cmdArgs,
	}

	err = commands.Run(state, cmd)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
