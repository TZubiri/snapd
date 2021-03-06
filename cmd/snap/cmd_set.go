// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package main

import (
	"fmt"
	"strings"

	"github.com/jessevdk/go-flags"

	"github.com/snapcore/snapd/i18n"
	"github.com/snapcore/snapd/jsonutil"
)

var shortSetHelp = i18n.G("Change configuration options")
var longSetHelp = i18n.G(`
The set command changes the provided configuration options as requested.

    $ snap set snap-name username=frank password=$PASSWORD

All configuration changes are persisted at once, and only after the
snap's configuration hook returns successfully.

Nested values may be modified via a dotted path:

    $ snap set snap-name author.name=frank

Configuration option may be unset with exclamation mark:
    $ snap set snap-name author!
`)

type cmdSet struct {
	waitMixin
	Positional struct {
		Snap       installedSnapName
		ConfValues []string `required:"1"`
	} `positional-args:"yes" required:"yes"`
}

func init() {
	addCommand("set", shortSetHelp, longSetHelp, func() flags.Commander { return &cmdSet{} }, waitDescs, []argDesc{
		{
			name: "<snap>",
			// TRANSLATORS: This should not start with a lowercase letter.
			desc: i18n.G("The snap to configure (e.g. hello-world)"),
		}, {
			// TRANSLATORS: This needs to begin with < and end with >
			name: i18n.G("<conf value>"),
			// TRANSLATORS: This should not start with a lowercase letter.
			desc: i18n.G("Set (key=value) or unset (key!) configuration value"),
		},
	})
}

func (x *cmdSet) Execute(args []string) error {
	patchValues := make(map[string]interface{})
	for _, patchValue := range x.Positional.ConfValues {
		parts := strings.SplitN(patchValue, "=", 2)
		if len(parts) == 1 && strings.HasSuffix(patchValue, "!") {
			patchValues[strings.TrimSuffix(patchValue, "!")] = nil
			continue
		}
		if len(parts) != 2 {
			return fmt.Errorf(i18n.G("invalid configuration: %q (want key=value)"), patchValue)
		}
		var value interface{}
		if err := jsonutil.DecodeWithNumber(strings.NewReader(parts[1]), &value); err != nil {
			// Not valid JSON-- just save the string as-is.
			patchValues[parts[0]] = parts[1]
		} else {
			patchValues[parts[0]] = value
		}
	}

	snapName := string(x.Positional.Snap)
	id, err := x.client.SetConf(snapName, patchValues)
	if err != nil {
		return err
	}

	if _, err := x.wait(id); err != nil {
		if err == noWait {
			return nil
		}
		return err
	}

	return nil
}
