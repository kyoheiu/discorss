package init

import (
	feed "github.com/kyoheiu/discorss/dfeed"

	_ "time/tzdata"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

func init() {
	functions.HTTP("SendFeed", feed.SendFeed)
}

// func main() {
// 	port := "8080"
// 	if envPort := os.Getenv("PORT"); envPort != "" {
// 		port = envPort
// 	}
// 	if err := funcframework.Start(port); err != nil {
// 		log.Fatalf("funcframework.Start: %v\n", err)
// 	}
// }
