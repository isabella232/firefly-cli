// Copyright © 2021 Kaleido, Inc.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package erc20

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/hyperledger/firefly-cli/internal/blockchain/ethereum"
	"github.com/hyperledger/firefly-cli/internal/constants"
	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/hyperledger/firefly-cli/pkg/types"
)

var TOKEN_INIT_ARGS = map[string]string{
	"name":   "initName",
	"symbol": "initSymbol",
}

func DeployContracts(s *types.Stack, log log.Logger, verbose bool) error {
	var containerName string
	for _, member := range s.Members {
		if !member.External {
			containerName = fmt.Sprintf("%s_tokens_%s", s.Name, member.ID)
			break
		}
	}
	if containerName == "" {
		return errors.New("unable to extract contracts from container - no valid tokens containers found in stack")
	}
	log.Info("extracting smart contracts")

	if err := ethereum.ExtractContracts(s.Name, containerName, "/root/contracts", verbose); err != nil {
		return err
	}

	// Token Factory
	tokenFactoryContract, err := ethereum.ReadCompiledContract(filepath.Join(constants.StacksDir, s.Name, "contracts", "ERC20WithDataFactory.json"))
	if err != nil {
		return err
	}
	var tokenFactoryContractAddress string
	for _, member := range s.Members {
		if tokenFactoryContractAddress == "" {
			// TODO: version the registered name
			log.Info(fmt.Sprintf("deploying ERC20 factory contract on '%s'", member.ID))
			tokenFactoryContractAddress, _, err = ethereum.DeployContract(member, tokenFactoryContract, "erc20Factory", TOKEN_INIT_ARGS)
			if err != nil {
				return err
			}
		} else {
			log.Info(fmt.Sprintf("registering ERC20 factory contract on '%s'", member.ID))
			err = ethereum.RegisterContract(member, tokenFactoryContract, tokenFactoryContractAddress, "erc20Factory", TOKEN_INIT_ARGS)
			if err != nil {
				return err
			}
		}
	}

	// Token contract
	tokenContract, err := ethereum.ReadCompiledContract(filepath.Join(constants.StacksDir, s.Name, "contracts", "ERC20WithData.json"))
	if err != nil {
		return err
	}
	var tokenContractAddress string
	var contractAbiID string
	for _, member := range s.Members {
		if tokenContractAddress == "" {
			log.Info(fmt.Sprintf("deploying ERC20 contract on '%s'", member.ID))
			tokenContractAddress, contractAbiID, err = ethereum.DeployContract(member, tokenContract, "erc20", TOKEN_INIT_ARGS)
			if err != nil {
				return err
			}
		} else {
			log.Info(fmt.Sprintf("registering ERC20 contract on '%s'", member.ID))
			err = ethereum.RegisterContract(member, tokenContract, tokenContractAddress, "erc20", TOKEN_INIT_ARGS)
			if err != nil {
				return err
			}
		}
	}
	// TODO: Use value as env variable
	fmt.Printf("\nERC20WithData.sol ABI ID: %s", contractAbiID)

	return nil
}
