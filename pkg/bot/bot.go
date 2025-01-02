package bot

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Replace with your own channel ID or forum channel ID.
// If using a forum, you may need to iterate over thread IDs instead.
const channelID = "1017265769102450748"

func Run(token string) {

	if token == "" {
		log.Fatal("Please set your DISCORD_BOT_TOKEN environment variable.")
	}

	// Create a new Discord session.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v\n", err)
	}

	// Open the WebSocket and begin listening.
	err = dg.Open()
	if err != nil {
		log.Fatalf("Error opening Discord session: %v\n", err)
	}
	defer dg.Close()

	log.Println("Bot is now running. Press CTRL-C to exit.")

	// Register the slash command during startup (optional: you can register once, or whenever you start up).
	registerSlashCommands(dg)

	// Add a handler for interaction events (slash commands).
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommand {
			switch i.ApplicationCommandData().Name {
			case "randomimage":
				handleRandomImage(s, i)
			}
		}
	})

	// Keep the program running until CTRL-C or an error occurs.
	select {}
}

// registerSlashCommands creates (and overwrites) the /randomimage command in your guild (or globally).
func registerSlashCommands(s *discordgo.Session) {
	_, err := s.ApplicationCommandCreate(
		s.State.User.ID,
		os.Getenv("GUILD"), // If empty, it registers globally. Otherwise, put a specific Guild ID to limit scope.
		&discordgo.ApplicationCommand{
			Name:        "randomimage",
			Description: "Returns a random image from a designated channel.",
		},
	)
	if err != nil {
		log.Printf("Cannot create slash command: %v\n", err)

	}
}

// handleRandomImage fetches recent messages from a specified channel, filters out image attachments,
// selects one at random, and sends it back in the slash command response.
func handleRandomImage(s *discordgo.Session, i *discordgo.InteractionCreate) {
	imageURL, err := getRandomImageURL(s, channelID)
	if err != nil {
		log.Printf("Error getting random image: %v", err)
		respondWithMessage(s, i, "Failed to find an image. Please try again or add some images!")
		return
	}

	respondWithMessage(s, i, fmt.Sprintf("Here is your random image:\n%s", imageURL))
}

// getRandomImageURL fetches messages in the channel, grabs attachments that are likely images, and picks one at random.
func getRandomImageURL(s *discordgo.Session, channelID string) (string, error) {
	// Fetch the most recent 100 messages (max allowed by Discord in one request).
	messages, err := s.ChannelMessages(channelID, 100, "", "", "")
	if err != nil {
		return "", fmt.Errorf("could not retrieve messages: %w", err)
	}

	var imageURLs []string
	for _, msg := range messages {
		// Check for attachments
		for _, attachment := range msg.Attachments {
			// You could also check the content type or extension here for more robust filtering.
			if isImageAttachment(attachment) {
				imageURLs = append(imageURLs, attachment.URL)
			}
		}

		// Optionally, if you want to include image links from message embeds:
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

	// Pick a random image from the slice
	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(imageURLs))
	return imageURLs[randomIndex], nil
}

// isImageAttachment is a basic check; you may want to refine this further.
func isImageAttachment(attachment *discordgo.MessageAttachment) bool {
	// Check file extension or ContentType if available.
	return attachment.Width > 0 && attachment.Height > 0
}

// respondWithMessage is a helper function to send a response to a slash command.
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
