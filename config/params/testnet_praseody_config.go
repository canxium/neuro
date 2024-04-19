package params

import "math"

// UsePraseOdyNetworkConfig uses the PraseOdy beacon chain specific network config.
func UsePraseOdyNetworkConfig() {
	cfg := BeaconNetworkConfig().Copy()
	cfg.ContractDeploymentBlock = 0
	cfg.BootstrapNodes = []string{
		// lighthouse
		"enr:-Le4QGgkMit_S6dcseu-6HUrQLKgXsOqjGmXvALew2xjINBwFeAGMGufIfOLQZoDXEBhqAnhjMW6eXufW81mVUNnnaMChGV0aDKQ2jJYTwAyMwABAAAAAAAAAIJpZIJ2NIJpcIQFobgAg2lwNpAqAQT_APCN3QAAAAAAAAABiXNlY3AyNTZrMaECNQgaxqD0o_mb6YMG5YDwIm0GhXMJF0iEuvAb-Sjl9KKDdWRwgiOMhHVkcDaCI-Y", // 40
		"enr:-Le4QJFLssDaSa99PqSFZVmzYPLFjLR0Glf5sLtNhluY1p4gV0NG1AWNiluckVx4ZNq7QCB_q-RbNUSvjZZ-wJl6XzYBhGV0aDKQ2jJYTwAyMwABAAAAAAAAAIJpZIJ2NIJpcIQxDY3Wg2lwNpAqAQT4wBI7yQAAAAAAAAABiXNlY3AyNTZrMaEDAnVP2fJn3ZQU2NBaoPOWYtKl7_JP1-6AMrZkMYM-x4CDdWRwgiOMhHVkcDaCI-Y", // 14
		"enr:-Le4QApx4yDdnYodma6kKim7v_UARV1B1AZ6fBMoMX2tM_gJQzOaY_lGfJ-e_q_3ZqZGavyIfhWme5SmdQTipYNa1bMBhGV0aDKQ2jJYTwAyMwABAAAAAAAAAIJpZIJ2NIJpcISyEInRg2lwNpAqAkeAABIjpwAAAAAAAAABiXNlY3AyNTZrMaEDCUfXoP8UZUHaVQpm8TQRhm-xWGtyHbAsi5hZxmUinhGDdWRwgiOMhHVkcDaCI-Y", // 09

		// lodestar
		"enr:-KG4QBhzrzGIi3fJCa3-GXBkvPBQXft1Gwmutwt3AtqIgje5MLoLyCDY0IIvxnqwE-F7GBvJUjMRl3x9zxbFgn6em3ECgmlkgnY0gmlwhAWhuACDaXA2kCoBBP8A8I3dAAAAAAAAAAGJc2VjcDI1NmsxoQO9vb9FwSJh3mxz3wTIdjpBCJr2NMlDnLQ_h3BOOJGQfoN1ZHCCI42EdWRwNoIj5w", // 51
		"enr:-KG4QCYl1Zl_XSN4_ly7n-5867u2ZbSGamFgawB1iI7-qkONGqFCyOC1XeVJEYlYoZ9oBRLfBwBAsIlloRWuHLU1A3kBgmlkgnY0gmlwhDENjdaDaXA2kCoBBPjAEjvJAAAAAAAAAAGJc2VjcDI1NmsxoQKkjFpo0r-D_6Cqnx3L-Mr1YdZc0mzakPF57hYWyz6DNIN1ZHCCI42EdWRwNoIj5w", // 49
		"enr:-KG4QDWqJ8wxnw4JHqQzxVSGSet4q8osBH568XkMnC1u8fPnPFARnkPIAq8KKEa8JKi6kN1EFSFkfnv8k6ZeWJPXTSkBgmlkgnY0gmlwhLIQidGDaXA2kCoCR4AAEiOnAAAAAAAAAAGJc2VjcDI1NmsxoQOEVvdJhjDQB6zwUNjYCNTvPXtCSKc2UqW5x2VoG5Gp04N1ZHCCI42EdWRwNoIj5w", // 17

		// static nodes
		"enr:-Nm4QHzN2cYmgnYqrK7dCNNnP44QP328uwJ0sWjLlDGNAhY4MfeaeEmFJ9kLeXY2-AYlfx8cwHTcjjbL1Cl_5AnqGXsFh2F0dG5ldHOIAAAAAIABAACEZXRoMpCceb3AADJTAP__________gmlkgnY0gmlwhA_rjYiDaXA2kCQCHwCAAAgAAAAAAAAALNyJc2VjcDI1NmsxoQNHAqh0K2EFoHK18tyrGNUSlLWs-5I93z2lfsXbEQd25ohzeW5jbmV0c4gAAAAAAAAAAIN0Y3CCIy6DdWRwgiMu", // lodestar
		"enr:-MK4QDZ9Vg_ty6msUFWiQC8aT4U_rZBS3Y3HPz5YfokuX6ipA0Mx6sqdn6-5vE1CMeS3GnFAHdvvhsd5SDLKB-vNgdKGAY70dagYh2F0dG5ldHOIAAAAAAAAAACEZXRoMpCceb3AADJjAP__________gmlkgnY0gmlwhMMjLZuJc2VjcDI1NmsxoQP9fEckcrTin5d_WCq0cBy5RurdNFCC-Z42IjDUlKvwRYhzeW5jbmV0cwCDdGNwgjLKg3VkcIIu4g",                               // prysm
	}
	OverrideBeaconNetworkConfig(cfg)
}

// PraseOdyConfig defines the config for the Sepolia beacon chain testnet.
func PraseOdyConfig() *BeaconChainConfig {
	cfg := MainnetConfig().Copy()
	cfg.ConfigName = PraseOdyName

	cfg.MinGenesisTime = 1712319000
	cfg.MinGenesisActiveValidatorCount = 164

	cfg.GenesisDelay = 3600

	cfg.DepositChainID = 30203
	cfg.DepositNetworkID = 30203
	cfg.DepositContractAddress = "0x55155Ca1F57bbDB1e8e10EBB871c00D809E5E84f"

	cfg.GenesisForkVersion = []byte{0x00, 0x32, 0x23, 0x00}
	cfg.AltairForkEpoch = 1
	cfg.AltairForkVersion = []byte{0x00, 0x32, 0x33, 0x00}
	cfg.BellatrixForkEpoch = 10
	cfg.BellatrixForkVersion = []byte{0x00, 0x32, 0x43, 0x00}
	cfg.CapellaForkEpoch = 40
	cfg.CapellaForkVersion = []byte{0x00, 0x32, 0x53, 0x00}
	cfg.DenebForkEpoch = math.MaxUint64
	cfg.DenebForkVersion = []byte{0x00, 0x32, 0x63, 0x00}

	cfg.TerminalTotalDifficulty = "1016000000"

	cfg.SecondsPerSlot = 6
	cfg.SecondsPerETH1Block = 6
	cfg.Eth1FollowDistance = 128

	cfg.BaseRewardFactor = 0
	cfg.BasePenaltyFactor = 0
	cfg.InactivityPenaltyQuotientBellatrix = 1 << 24
	cfg.MaxExcessBalance = 32 * 1e10

	cfg.InitializeForkSchedule()
	return cfg
}
