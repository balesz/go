package cmd

import (
	"context"
	"log"

	"firebase.google.com/go/v4/auth"

	"github.com/balesz/go/firebase"

	"github.com/spf13/cobra"
)

func init() {
	log.SetFlags(log.Flags() &^ log.Ldate &^ log.Ltime)

	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(updateCmd, deleteCmd)

	authCmd.PersistentFlags().String("uid", "", "User unique identifier")
	authCmd.PersistentFlags().String("email", "", "User email address")

	updateCmd.Flags().String("displayName", "", "Display name")
}

var authCmd = &cobra.Command{Use: "auth", Short: "Authentication commands"}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete user",
	PreRun: func(cmd *cobra.Command, args []string) {
		checkEnvironment()
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		if uid := cmd.Flag("uid").Value.String(); uid != "" {
			if err := firebase.Auth.DeleteUser(ctx, uid); err != nil {
				log.Fatalln(err)
			}
		} else if email := cmd.Flag("email").Value.String(); email != "" {
			if rec, err := firebase.Auth.GetUserByEmail(ctx, email); err != nil {
				log.Fatalln(err)
			} else if err := firebase.Auth.DeleteUser(ctx, rec.UID); err != nil {
				log.Fatalln(err)
			}
		} else {
			log.Fatalf("uid or email parameter is required")
		}
	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update user",
	PreRun: func(cmd *cobra.Command, args []string) {
		checkEnvironment()
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		var update = auth.UserToUpdate{}
		if name := cmd.Flag("displayName").Value.String(); name != "" {
			update.DisplayName(name)
		}

		var err error
		var rec *auth.UserRecord

		if uid := cmd.Flag("uid").Value.String(); uid != "" {
			if rec, err = firebase.Auth.GetUser(ctx, uid); err != nil {
				log.Fatalln(err)
			}
		} else if email := cmd.Flag("email").Value.String(); email != "" {
			if rec, err = firebase.Auth.GetUserByEmail(ctx, email); err != nil {
				log.Fatalln(err)
			}
		} else {
			log.Fatalf("uid or email parameter is required")
		}

		if rec, err = firebase.Auth.UpdateUser(ctx, rec.UID, &update); err != nil {
			log.Fatalln(err)
		}

		log.Println(rec)
	},
}

func checkEnvironment() {
	if err := firebase.CheckEnvironment(); err != nil {
		log.Fatalln(err)
	}
}
