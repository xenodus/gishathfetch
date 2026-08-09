/**
 * Keep in sync with frontend/src/constants.js (store list and search limits).
 * Long term: move to packages/shared/.
 */

const AGORA_STORE_NAME = "Agora Hobby";
const AGORA_SEARCH_ENABLED = true;

const ALL_LGS_OPTIONS = [
  "5 Mana",
  AGORA_STORE_NAME,
  "Arcane Sanctum",
  "Card Affinity",
  "Cards & Collections",
  "Cards Central",
  "Cards Citadel",
  "Dueller's Point",
  "Flagship Games",
  "Fyendal Hobby",
  "Games Haven",
  "Grey Ogre Games",
  "Hideout",
  "Hideyoshi",
  "Mana Pro",
  "Mox & Lotus",
  "MTG Asia",
  "OneMtg",
  "The TCG Marketplace",
];

export const LGS_OPTIONS = AGORA_SEARCH_ENABLED
  ? ALL_LGS_OPTIONS
  : ALL_LGS_OPTIONS.filter((store) => store !== AGORA_STORE_NAME);

export const SITE_TAGLINE =
  "Magic: The Gathering price checker for Singapore's LGS and online shops";

export const MIN_SEARCH_LENGTH = 3;
export const MAX_SEARCH_LENGTH = 150;

export type StoreLocation = {
  id: string;
  name: string;
  address: string;
  website: string;
};

/** Store locations for the map tab (no iframe embeds on native). */
export const LGS_MAP: StoreLocation[] = [
  {
    id: "5-mana-map",
    name: "5 Mana",
    address: "511 Guillemard Rd, #02-06, Singapore 399849",
    website: "https://5-mana.sg/",
  },
  {
    id: "agora-map",
    name: "Agora Hobby",
    address: "French Rd, #05-164 Blk 809, Singapore 200809",
    website: "https://agorahobby.com/",
  },
  {
    id: "arcane-sanctum-map",
    name: "Arcane Sanctum",
    address: "809 French Rd, #02-36 Kitchener Complex, Singapore 200809",
    website: "https://arcanesanctumtcg.com/",
  },
  {
    id: "cards-central-map",
    name: "Cards Central",
    address: "62A Smith Street, Chinatown, Singapore 058964",
    website: "https://cardscentral.com/",
  },
  {
    id: "cards-citadel-map",
    name: "Cards Citadel",
    address: "Blk 10 North Bridge Road, #02-5117, Singapore 190010",
    website: "https://cardscitadel.com/",
  },
  {
    id: "dueller-point-map",
    name: "Dueller's Point",
    address: "450 Hougang Ave 10, B1-541, Singapore 530450",
    website: "https://www.duellerspoint.com/",
  },
  {
    id: "flagship-games-map",
    name: "Flagship Games",
    address: "214 Bishan St. 23, B1-223, Singapore 570214",
    website: "https://www.flagshipgames.sg/",
  },
  {
    id: "fyendal-hobby-map",
    name: "Fyendal Hobby",
    address: "86 Marine Parade Central, #04-305, Singapore 440086",
    website: "https://fyendalhobby.com/",
  },
  {
    id: "games-haven-pl-map",
    name: "Games Haven - Paya Lebar",
    address: "736 Geylang Rd, Singapore 389647",
    website: "https://www.gameshaventcg.com/",
  },
  {
    id: "grey-ogre-map",
    name: "Grey Ogre Games",
    address: "83 Club St, Singapore 069451",
    website: "https://www.greyogregames.com/",
  },
  {
    id: "hideyoshi-map",
    name: "Hideyoshi",
    address:
      "504 Jurong West Street 51, #04-211, Hong Kah Court, Singapore 640504",
    website: "https://hideyoshitcg.com/",
  },
  {
    id: "hideout-map",
    name: "Hideout",
    address: "803 King George's Ave, #02-190, Singapore 200803",
    website: "https://hideoutcg.com/",
  },
  {
    id: "manapro-map",
    name: "Mana Pro",
    address: "BLK 203 Choa Chu Kang Ave 1, B1-41, Singapore 680203",
    website: "https://sg-manapro.com/",
  },
  {
    id: "mox-map",
    name: "Mox & Lotus",
    address: "215 Bedok North Street 1, #02-85, Singapore 460215",
    website: "https://www.moxandlotus.sg/",
  },
  {
    id: "mtg-asia-map",
    name: "MTG Asia",
    address: "261 Waterloo St, #03-28, Singapore 180261",
    website: "https://www.mtg-asia.com/",
  },
  {
    id: "onemtg-map",
    name: "One MTG",
    address: "100 Jln Sultan, #03-11 Sultan Plaza, Singapore 199001",
    website: "https://onemtg.com.sg/",
  },
];
