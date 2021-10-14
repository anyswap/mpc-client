package utils

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

var (
	// LicenseCommand license cubcommonad
	LicenseCommand = &cli.Command{
		Action:    license,
		Name:      "license",
		Usage:     "Display license information",
		ArgsUsage: " ",
	}
)

func license(_ *cli.Context) error {
	fmt.Println(`MPC-Client is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

MPC-Client is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with anyswap. If not, see <http://www.gnu.org/licenses/>.`)
	return nil
}
