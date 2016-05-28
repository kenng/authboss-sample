package main

import "os"

func main() {
    if (os.Getenv("AUTHBOSS_FRAMEWORK") == "gin") {
        mainGin()
    } else {
        mainDefault()
    }
}
