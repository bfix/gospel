package wallet

//----------------------------------------------------------------------
// This file is part of Gospel.
// Copyright (C) 2011-2021 Bernd Fix  >Y<
//
// Gospel is free software: you can redistribute it and/or modify it
// under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License,
// or (at your option) any later version.
//
// Gospel is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//
// SPDX-License-Identifier: AGPL3.0-or-later
//----------------------------------------------------------------------

// Address constants
const (
	// Mainnet/Testnet/Regnet
	AddrMain = 0
	AddrTest = 1
	AddrReg  = 2

	// Address usage
	AddrP2PKH        = 0
	AddrP2SH         = 1
	AddrP2WPKH       = 2
	AddrP2WSH        = 3
	AddrP2WPKHinP2SH = 4
	AddrP2WSHinP2SH  = 5
)

// AddrVersion defines address version constants
type AddrVersion struct {
	Version    uint16 // version byte (address prefix)
	PubVersion uint32 // BIP32 key version (public)
	PrvVersion uint32 // BIP32 key version (private)
}

// AddrFormat defines formatting information for addresses
type AddrFormat struct {
	Bech32     string
	WifVersion int
	Versions   []*AddrVersion
}

// AddrSpec defines a coin address format.
type AddrSpec struct {
	CoinID  int
	Formats []*AddrFormat
}

/*
	{2, "LTC", "Litecoin", 176, 48, 5, 0x0488ade4, 0x0488b21e, 1},
	{3, "DOGE", "Dogecoin", 158, 30, 22, 0x02fac398, 0x02facafd, 1},
*/

var (
	// AddrList for selected coins
	// (see page source for "https://iancoleman.io/bip39/")
	AddrList = []*AddrSpec{
		//--------------------------------------------------------------
		// BTC (Bitcoin)
		//--------------------------------------------------------------
		{0, []*AddrFormat{
			// Mainnet
			{"bc", 0x80, []*AddrVersion{
				{0x00, 0x0488ade4, 0x0488b21e}, // P2PKH
				{0x05, 0x0488ade4, 0x0488b21e}, // P2SH
				{0x00, 0x04b24746, 0x04b2430c}, // P2WPKH
				{0x05, 0x02aa7ed3, 0x02aa7a99}, // P2WSH
				{0x05, 0x049d7cb2, 0x049d7878}, // P2WPKHinP2SH
				{0x05, 0x0295b43f, 0x0295b005}, // P2WSHinP2SH
			}},
			// Testnet
			{"tb", 0xef, []*AddrVersion{
				{0x6f, 0x0488ade4, 0x0488b21e}, // P2PKH
				{0xc4, 0x0488ade4, 0x0488b21e}, // P2SH
				{0x6f, 0x045f1cf6, 0x045f18bc}, // P2WPKH
				{0xc4, 0x02575483, 0x02575048}, // P2WSH
				{0xc4, 0x044a5262, 0x044a4e28}, // P2WPKHinP2SH
				{0xc4, 0x024289ef, 0x024285b5}, // P2WSHinP2SH
			}},
			// Regnet
			{"bcrt", 0xef, []*AddrVersion{
				{0x6f, 0x043587cf, 0x04358394}, // P2PKH
				{0xc4, 0x043587cf, 0x04358394}, // P2SH
				{0x6f, 0x045f1cf6, 0x045f18bc}, // P2WPKH
				{0xc4, 0x02575483, 0x02575048}, // P2WSH
				{0xc4, 0x044a5262, 0x044a4e28}, // P2WPKHinP2SH
				{0xc4, 0x024289ef, 0x024285b5}, // P2WSHinP2SH
			}},
		}},
		//--------------------------------------------------------------
		// LTC (Litecoin)
		//--------------------------------------------------------------
		{2, []*AddrFormat{
			// Mainnet
			{"ltc", 0xb0, []*AddrVersion{
				{0x30, 0x0488ade4, 0x0488b21e}, // P2PKH
				{0x32, 0x0488ade4, 0x0488b21e}, // P2SH
				{0x30, 0x04b24746, 0x04b2430c}, // P2WPKH
				nil,                            // P2WSH
				{0x32, 0x01b26ef6, 0x01b26792}, // P2WPKHinP2SH
				nil,                            // P2WSHinP2SH
			}},
			// Testnet
			{"litecointestnet", 0xef, []*AddrVersion{
				{0x6f, 0x043587cf, 0x04358394}, // P2PKH
				{0xc4, 0x043587cf, 0x04358394}, // P2SH
				{0x6f, 0x043587cf, 0x04358394}, // P2WPKH
				nil,                            // P2WSH
				{0xc4, 0x043587cf, 0x04358394}, // P2WPKHinP2SH
				nil,                            // P2WSHinP2SH
			}},
			// Regnet
			nil,
		}},
		//--------------------------------------------------------------
		// DOGE
		//--------------------------------------------------------------
		{3, []*AddrFormat{
			// Mainnet
			{"", 0x9e, []*AddrVersion{
				{0x1e, 0x02facafd, 0x02fac398}, // P2PKH
				{0x16, 0x02facafd, 0x02fac398}, // P2SH
				nil,                            // P2WPKH
				nil,                            // P2WSH
				nil,                            // P2WPKHinP2SH
				nil,                            // P2WSHinP2SH
			}},
			// Testnet
			{"dogecointestnet", 0xf1, []*AddrVersion{
				{0x71, 0x043587cf, 0x04358394}, // P2PKH
				{0xc4, 0x043587cf, 0x04358394}, // P2SH
				{0x71, 0x043587cf, 0x04358394}, // P2WPKH
				nil,                            // P2WSH
				{0xc4, 0x043587cf, 0x04358394}, // P2WPKHinP2SH
				nil,                            // P2WSHinP2SH
			}},
			// Regnet
			nil,
		}},
		//--------------------------------------------------------------
		// DASH
		//--------------------------------------------------------------
		{5, []*AddrFormat{
			// Mainnet
			{"", 0xcc, []*AddrVersion{
				{0x4c, 0x0488ade4, 0x0488b21e}, // P2PKH
				{0x10, 0x0488ade4, 0x0488b21e}, // P2SH
				nil,                            // P2WPKH
				nil,                            // P2WSH
				nil,                            // P2WPKHinP2SH
				nil,                            // P2WSHinP2SH
			}},
			// Testnet
			{"", 0xef, []*AddrVersion{
				{0x8c, 0x043587cf, 0x04358394}, // P2PKH
				{0x13, 0x043587cf, 0x04358394}, // P2SH
				nil,                            // P2WPKH
				nil,                            // P2WSH
				nil,                            // P2WPKHinP2SH
				nil,                            // P2WSHinP2SH
			}},
			// Regnet
			nil,
		}},
		//--------------------------------------------------------------
		// NMC (Namecoin)
		//--------------------------------------------------------------
		{7, []*AddrFormat{
			// Mainnet
			{"", 0xb4, []*AddrVersion{
				{0x34, 0x0488ade4, 0x0488b21e}, // P2PKH
				{0x0d, 0x0488ade4, 0x0488b21e}, // P2SH
				nil,                            // P2WPKH
				nil,                            // P2WSH
				nil,                            // P2WPKHinP2SH
				nil,                            // P2WSHinP2SH
			}},
			// Testnet
			nil,
			// Regnet
			nil,
		}},
		//--------------------------------------------------------------
		// DGB (Digibyte)
		//--------------------------------------------------------------
		{20, []*AddrFormat{
			// Mainnet
			{"dgb", 0x80, []*AddrVersion{
				{0x1e, 0x0488ade4, 0x0488b21e}, // P2PKH
				{0x05, 0x0488ade4, 0x0488b21e}, // P2SH
				{0x1e, 0x04b24746, 0x04b2430c}, // P2WPKH
				nil,                            // P2WSH
				{0x3f, 0x049d7cb2, 0x049d7878}, // P2WPKHinP2SH
				nil,                            // P2WSHinP2SH
			}},
			// Testnet
			nil,
			// Regnet
			nil,
		}},
		//--------------------------------------------------------------
		// ZEC (ZCash)
		//--------------------------------------------------------------
		{133, []*AddrFormat{
			// Mainnet
			{"", 0x80, []*AddrVersion{
				{0x1cb8, 0x0488ade4, 0x0488b21e}, // P2PKH
				{0x1cbd, 0x0488ade4, 0x0488b21e}, // P2SH
				nil,                              // P2WPKH
				nil,                              // P2WSH
				nil,                              // P2WPKHinP2SH
				nil,                              // P2WSHinP2SH
			}},
			// Testnet
			nil,
			// Regnet
			nil,
		}},
		//--------------------------------------------------------------
		// BCH
		//--------------------------------------------------------------
		{145, []*AddrFormat{
			// Mainnet
			{"", 0x80, []*AddrVersion{
				{0x00, 0x0488ade4, 0x0488b21e}, // P2PKH
				{0x05, 0x0488ade4, 0x0488b21e}, // P2SH
				{0x00, 0x04b24746, 0x04b2430c}, // P2WPKH
				{0x05, 0x02aa7ed3, 0x02aa7a99}, // P2WSH
				{0x05, 0x049d7cb2, 0x049d7878}, // P2WPKHinP2SH
				{0x05, 0x0295b43f, 0x0295b005}, // P2WSHinP2SH
			}},
			// Testnet
			{"", 0xef, []*AddrVersion{
				{0x6f, 0x0488ade4, 0x0488b21e}, // P2PKH
				{0xc4, 0x0488ade4, 0x0488b21e}, // P2SH
				{0x6f, 0x045f1cf6, 0x045f18bc}, // P2WPKH
				{0xc4, 0x02575483, 0x02575048}, // P2WSH
				{0xc4, 0x044a5262, 0x044a4e28}, // P2WPKHinP2SH
				{0xc4, 0x024289ef, 0x024285b5}, // P2WSHinP2SH
			}},
			// Regnet
			{"", 0xef, []*AddrVersion{
				{0x6f, 0x043587cf, 0x04358394}, // P2PKH
				{0xc4, 0x043587cf, 0x04358394}, // P2SH
				{0x6f, 0x045f1cf6, 0x045f18bc}, // P2WPKH
				{0xc4, 0x02575483, 0x02575048}, // P2WSH
				{0xc4, 0x044a5262, 0x044a4e28}, // P2WPKHinP2SH
				{0xc4, 0x024289ef, 0x024285b5}, // P2WSHinP2SH
			}},
		}},
		//--------------------------------------------------------------
		// BTG
		//--------------------------------------------------------------
		{156, []*AddrFormat{
			// Mainnet
			{"btg", 0x80, []*AddrVersion{
				{0x26, 0x0488ade4, 0x0488b21e}, // P2PKH
				{0x17, 0x0488ade4, 0x0488b21e}, // P2SH
				{0x26, 0x04b24746, 0x04b2430c}, // P2WPKH
				nil,                            // P2WSH
				{0x17, 0x049d7cb2, 0x049d7878}, // P2WPKHinP2SH
				nil,                            // P2WSHinP2SH
			}},
			// Testnet
			nil,
			// Regnet
			nil,
		}},
	}
)
