package processors

import (
	"strings"

	"github.com/nanobox-io/nanobox/commands/registry"
	"github.com/nanobox-io/nanobox/helpers"
	"github.com/nanobox-io/nanobox/models"
	"github.com/nanobox-io/nanobox/util"
	"github.com/nanobox-io/nanobox/util/config"
	"github.com/nanobox-io/nanobox/util/nanoagent"
	"github.com/nanobox-io/nanobox/util/odin"
)

func Console(envModel *models.Env, consoleConfig ConsoleConfig) error {

	appID := consoleConfig.App

	// fetch the remote
	remote, ok := envModel.Remotes[appID]
	if ok {
		// set the odin endpoint
		odin.SetEndpoint(remote.Endpoint)
		// set the app id
		appID = remote.ID
	}

	// set odins endpoint if the arguement is passed
	if endpoint := registry.GetString("endpoint"); endpoint != "" {
		odin.SetEndpoint(endpoint)
	}

	// set the app id to the directory name if it's default
	if appID == "default" {
		appID = config.AppName()
	}

	// validate access to the app
	if err := helpers.ValidateOdinApp(appID); err != nil {
		return util.ErrorAppend(err, "unable to validate app")
	}

	// initiate a console session with odin
	key, location, protocol, err := odin.EstablishConsole(appID, consoleConfig.Host)
	if err != nil {
		// todo: can we know if the request was rejected for authorization and print that?
		// We may not want that^ as it introduces the potential for app enumeration
		err = util.ErrorAppend(err, "failed to initiate a remote console session")
		if err != nil {
			if err2, ok := err.(util.Err); ok {
				if strings.Contains(err2.Error(), "Internal Server Error") {
					err2.Suggest = "It appears there was an issue with the request. If subsequent attempts fail, please report."
				} else {
					err2.Suggest = "It appears there is no component/host by that name, check the component/host name and try again"
				}
				return err2
			}
		}
		return err
	}

	switch protocol {
	case "docker":
		if err := nanoagent.Console(key, location); err != nil {
			return util.ErrorAppend(err, "failed to connect to remote console session")
		}
	case "ssh":
		if err := nanoagent.SSH(key, location); err != nil {
			return util.ErrorAppend(err, "failed to connect to remote ssh server")
		}
	}

	return nil
}
