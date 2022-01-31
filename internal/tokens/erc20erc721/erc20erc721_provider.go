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

package erc20erc721

import (
	"fmt"

	"github.com/hyperledger/firefly-cli/internal/core"
	"github.com/hyperledger/firefly-cli/internal/docker"
	"github.com/hyperledger/firefly-cli/internal/log"
	"github.com/hyperledger/firefly-cli/pkg/types"
)

type ERC20ERC721Provider struct {
	Log     log.Logger
	Verbose bool
	Stack   *types.Stack
}

func (p *ERC20ERC721Provider) DeploySmartContracts() error {
	return nil
}

func (p *ERC20ERC721Provider) FirstTimeSetup() error {
	for _, member := range p.Stack.Members {
		p.Log.Info(fmt.Sprintf("initializing tokens on member %s", member.ID))
		tokenInitUrl := fmt.Sprintf("http://localhost:%d/api/v1/init", member.ExposedTokensPort)
		if err := core.RequestWithRetry("POST", tokenInitUrl, nil, nil); err != nil {
			return err
		}
	}
	return nil
}

func (p *ERC20ERC721Provider) GetDockerServiceDefinitions() []*docker.ServiceDefinition {
	serviceDefinitions := make([]*docker.ServiceDefinition, 0, len(p.Stack.Members))
	for i, member := range p.Stack.Members {
		serviceDefinitions = append(serviceDefinitions, &docker.ServiceDefinition{
			ServiceName: "tokens_" + member.ID,
			Service: &docker.Service{
				Image:         p.Stack.VersionManifest.TokensERC20ERC721.GetDockerImageString(),
				ContainerName: fmt.Sprintf("%s_tokens_%v", p.Stack.Name, i),
				Ports:         []string{fmt.Sprintf("%d:3000", member.ExposedTokensPort)},
				Environment: map[string]string{
					"ETHCONNECT_CONTRACT_URI": "/abis/",
					"ETHCONNECT_INSTANCE":     "/contracts/erc20erc721Factory",
					"ETHCONNECT_TOPIC":        "tokensERC20ERC721",
					"ETHCONNECT_URL":          p.getEthconnectURL(member),
					"AUTO_INIT":               "false",
				},
				DependsOn: map[string]map[string]string{
					"ethconnect_" + member.ID: {"condition": "service_started"},
				},
				HealthCheck: &docker.HealthCheck{
					Test: []string{"CMD", "curl", "http://localhost:3000/api"},
				},
				Logging: docker.StandardLogOptions,
			},
		})
	}
	return serviceDefinitions
}

func (p *ERC20ERC721Provider) GetFireflyConfig(m *types.Member) *core.TokensConfig {
	return &core.TokensConfig{
		&core.TokenConnector{
			Plugin: "fftokens",
			Name:   "erc20erc721",
			URL:    p.getTokensURL(m),
		},
	}
}

func (p *ERC20ERC721Provider) getEthconnectURL(member *types.Member) string {
	return fmt.Sprintf("http://ethconnect_%s:8080", member.ID)
}

func (p *ERC20ERC721Provider) getTokensURL(member *types.Member) string {
	if !member.External {
		return fmt.Sprintf("http://tokens_%s:3000", member.ID)
	} else {
		return fmt.Sprintf("http://127.0.0.1:%v", member.ExposedTokensPort)
	}
}
