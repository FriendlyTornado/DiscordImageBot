package bot

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
)

// channel ID or forum channel ID.
const channelID = ""

func Run(token string) {

	if token == "" {
		log.Fatal("Please set your DISCORD_BOT_TOKEN environment variable.")
	}

	// Create a new Discord session.
	botSession, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v\n", err)
	}

	err = botSession.Open()
	if err != nil {
		log.Fatalf("Error opening Discord session: %v\n", err)
	}
	defer botSession.Close()

	log.Println("Bot is now running. Press CTRL-C to exit.")

	registerSlashCommands(botSession)

	// slash command handler
	botSession.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommand {
			switch i.ApplicationCommandData().Name {
			case "randomimage":
				handleRandomImage(s, i)
			}
		}
	})

	// Keeps bot running until ctrl-c or an error occurs.
	select {}
}

// creates /randomimage slash command in your guild
func registerSlashCommands(s *discordgo.Session) {
	_, err := s.ApplicationCommandCreate(
		s.State.User.ID,
		os.Getenv("GUILD"),
		&discordgo.ApplicationCommand{
			Name:        "randomimage",
			Description: "Returns a random image from a designated channel.",
		},
	)
	if err != nil {
		log.Printf("Cannot create slash command: %v\n", err)

	}
}

// fetches recent messages from a specified channel, filters out image attachments
func handleRandomImage(s *discordgo.Session, i *discordgo.InteractionCreate) {
	imageURL, err := getRandomImageURL(s, channelID)
	if err != nil {
		log.Printf("Error getting random image: %v", err)
		respondWithMessage(s, i, "Failed to find an image. Please try again or add some images!")
		return
	}

	respondWithMessage(s, i, fmt.Sprintf("Here is your random image:\n%s", imageURL))
}

// fetches messages in the channel, grabs attachments
func getRandomImageURL(s *discordgo.Session, channelID string) (string, error) {
	messages, err := s.ChannelMessages(channelID, 100, "", "", "")
	if err != nil {
		return "", fmt.Errorf("could not retrieve messages: %w", err)
	}

	var imageURLs []string
	for _, msg := range messages {
		for _, attachment := range msg.Attachments {
			if isImageAttachment(attachment) {
				imageURLs = append(imageURLs, attachment.URL)
			}
		}

		for _, embed := range msg.Embeds {
			if embed.Type == discordgo.EmbedTypeImage && embed.URL != "" {
				imageURLs = append(imageURLs, embed.URL)
			} else if embed.Image != nil && embed.Image.URL != "" {
				imageURLs = append(imageURLs, embed.Image.URL)
			}
		}
	}

	if len(imageURLs) == 0 {
		return "", fmt.Errorf("no image attachments found in channel")
	}

	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(imageURLs))
	return imageURLs[randomIndex], nil
}

// file filter
func isImageAttachment(attachment *discordgo.MessageAttachment) bool {
	// Check file extension or ContentType if available.
	return attachment.Width > 0 && attachment.Height > 0
}

// helper function to send a response to a slash command.
func respondWithMessage(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
	if err != nil {
		log.Printf("Error responding to interaction: %v", err)
	}
}
