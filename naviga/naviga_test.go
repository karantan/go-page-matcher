// Source from https://github.com/YoungiiJC/go-rod-aws-lambda/blob/main/rod.go
package naviga

import (
	"os"
	"page_matcher/utils"
	"testing"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	godotenv.Load(utils.RootDir() + "/.env")
	code := m.Run()
	os.Exit(code)
}

// func TestTakeScreenshot(t *testing.T) {
// 	defer os.Remove(utils.RootDir() + "/lib/wikipedia.org.png")

// 	err := TakeScreenshot("https://wikipedia.org", utils.RootDir()+"/lib")
// 	assert.FileExists(t, utils.RootDir()+"/lib/wikipedia.org.png")
// 	assert.NoError(t, err)
// }
