// --- App Constants ---
export const PAGE_TITLE = "Gishath Fetch: MTG Price Checker for Singapore's LGS";

export const LGS_OPTIONS = [
    "5 Mana",
    "Agora Hobby",
    "Arcane Sanctum",
    "Card Affinity",
    "Cardboard Crack Games",
    "Cards Citadel",
    "Cards & Collections",
    "Dueller's Point",
    "Flagship Games",
    "Games Haven",
    "Grey Ogre Games",
    "Hideout",
    "Mana Pro",
    "Mox & Lotus",
    "MTG Asia",
    "OneMtg",
    "Tefuda",
    "The TCG Marketplace"
];

export const LGS_MAP = [
    {
        id: "5-mana-map",
        name: "5 Mana",
        address: "511 Guillemard Rd, #02-06, Singapore 399849",
        iframe: "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d3988.7686522542544!2d103.88875987494157!3d1.314306298673231!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da19ef83dc5edf%3A0xf45523d5c3efb509!2s5-MANA.SG!5e0!3m2!1sen!2ssg!4v1768142747318!5m2!1sen!2ssg",
        website: "https://5-mana.sg/"
    },
    {
        id: "agora-map",
        name: "Agora Hobby",
        address: "French Rd, #05-164 Blk 809, Singapore 200809",
        iframe: "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d3988.778050505021!2d103.85967687451628!3d1.3084089617085968!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da19c9f7d7f74d%3A0xeaa1a66df7d4bcd6!2sAgora%20Hobby!5e0!3m2!1sen!2ssg!4v1702820213937!5m2!1sen!2ssg",
        website: "https://agorahobby.com/"
    },
    {
        id: "arcane-sanctum-map",
        name: "Arcane Sanctum",
        address: "809 French Rd, #02-36 Kitchener Complex, Singapore 200809",
        iframe: "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d3988.778059032544!2d103.8596768749415!3d1.3084035986791807!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da197e8c761a49%3A0x8c56b7150064528b!2sArcane%20Sanctum!5e0!3m2!1sen!2ssg!4v1768317836907!5m2!1sen!2ssg",
        website: "https://arcanesanctumtcg.com/"
    },
    {
        id: "cardboard-crack-games-map",
        name: "Cardboard Crack Games",
        address: "Upper Bukit Timah Rd, #03-28 Beauty World Centre, 144, Singapore 588177",
        iframe: "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d3988.7233430292667!2d103.7736843749657!3d1.3423740986449086!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da1920d676db93%3A0xe7b298b897da7b52!2sCardboard%20Crack%20Games!5e0!3m2!1sen!2ssg!4v1731824736033!5m2!1sen!2ssg",
        website: "https://www.cardboardcrackgames.com/"
    },
    {
        id: "cards-citadel-map",
        name: "Cards Citadel",
        address: "464 Crawford Ln, #02-01, Singapore 190464",
        iframe: "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d3988.783678524258!2d103.85966947451631!3d1.3048646617197366!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da190c9e183751%3A0xa2119a95d1e683f2!2sCards%20Citadel!5e0!3m2!1sen!2ssg!4v1702820792196!5m2!1sen!2ssg",
        website: "https://cardscitadel.com/"
    },
    {
        id: "dueller-point-map",
        name: "Dueller's Point",
        address: "450 Hougang Ave 10, B1-541, Singapore 530450",
        iframe: "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d3988.662159756766!2d103.89300967451602!3d1.3793695614811952!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da163eecb250ff%3A0xc7c259e72671dc62!2sDueller&#39;s%20Point!5e0!3m2!1sen!2ssg!4v1702820876967!5m2!1sen!2ssg",
        website: "https://www.duellerspoint.com/"
    },
    {
        id: "flagship-games-map",
        name: "Flagship Games",
        address: "214 Bishan St. 23, B1-223, Singapore 570214",
        iframe: "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d490.0218996351789!2d103.84829838647084!3d1.3574649065942905!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da173ef6ffcc0b%3A0x880386dee363a253!2sFlagship%20Games!5e0!3m2!1sen!2ssg!4v1734958555684!5m2!1sen!2ssg",
        website: "https://www.flagshipgames.sg/"
    },
    {
        id: "games-haven-pl-map",
        name: "Games Haven - Paya Lebar",
        address: "736 Geylang Rd, Singapore 389647",
        iframe: "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d63819.358332241325!2d103.79905633083244!3d1.350592080054757!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da1817d10ac901%3A0x2cacb3a0679089a2!2sGames%20Haven!5e0!3m2!1sen!2ssg!4v1702821045126!5m2!1sen!2ssg",
        website: "https://www.gameshaventcg.com/"
    },
    {
        id: "grey-ogre-map",
        name: "Grey Ogre Games",
        address: "83 Club St, Singapore 069451",
        iframe: "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d3988.8199964760065!2d103.84085797576442!3d1.2817574584814586!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da190d242b70db%3A0x965b932c3bc19eda!2sGrey%20Ogre%20Games!5e0!3m2!1sen!2ssg!4v1702821297360!5m2!1sen!2ssg",
        website: "https://www.greyogregames.com/"
    },
    {
        id: "hideout-map",
        name: "Hideout",
        address: "803 King George's Ave, #02-190, Singapore 200803",
        iframe: "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d15955.112777358516!2d103.84179288715819!3d1.3083185!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da19e075f4c4f5%3A0x60e4a2c61816be63!2sHideout!5e0!3m2!1sen!2ssg!4v1702821327690!5m2!1sen!2ssg",
        website: "https://hideoutcg.com/"
    },
    {
        id: "manapro-map",
        name: "Mana Pro",
        address: "BLK 203 Choa Chu Kang Ave 1, B1-41, Singapore 680203",
        iframe: "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d3988.6584888121897!2d103.74693327451605!3d1.3815577614740542!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da1176665e2737%3A0x3b8608ab4d67724f!2sMana%20Pro!5e0!3m2!1sen!2ssg!4v1702821359528!5m2!1sen!2ssg",
        website: "https://sg-manapro.com/"
    },
    {
        id: "mox-map",
        name: "Mox & Lotus",
        address: "215 Bedok North Street 1, #02-85, Singapore 460215",
        iframe: "https://www.google.com/maps/embed?pb=!1m14!1m8!1m3!1d15954.999958678827!2d103.9334704!3d1.3259392!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da19d89d198d6b%3A0xb3e238feedd6c90d!2sMox%20%26%20Lotus!5e0!3m2!1sen!2ssg!4v1730797737444!5m2!1sen!2ssg",
        website: "https://www.moxandlotus.sg/"
    },
    {
        id: "mtg-asia-map",
        name: "MTG Asia",
        address: "261 Waterloo St, #03-28, Singapore 180261",
        iframe: "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d3988.7930896678654!2d103.8493947744998!3d1.2989162986887468!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da19bb4a2bee83%3A0x28725aa3a3e2a51!2sMTG-Asia!5e0!3m2!1sen!2ssg!4v1703085334392!5m2!1sen!2ssg",
        website: "https://www.mtg-asia.com/"
    },
    {
        id: "onemtg-map",
        name: "One MTG",
        address: "100 Jln Sultan, #03-11 Sultan Plaza, Singapore 199001",
        iframe: "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d3988.7866900551694!2d103.85910407451628!3d1.3029641617257042!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da19180d91f3a1%3A0x75c807bf93d430a4!2sOne%20MTG!5e0!3m2!1sen!2ssg!4v1702821425238!5m2!1sen!2ssg",
        website: "https://onemtg.com.sg/"
    },
    {
        id: "tefuda-map",
        name: "Tefuda",
        address: "B1-02 Macpherson Mall, Singapore 368125",
        iframe: "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d3988.740634996764!2d103.8765490749657!3d1.3317319986556433!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da178ea031acdb%3A0xac7ea94397d6a870!2sTefuda!5e0!3m2!1sen!2ssg!4v1743179304416!5m2!1sen!2ssg",
        website: "https://tefudagames.com/"
    },
    {
        id: "unsleeved-map",
        name: "Unsleeved",
        address: "17A Jln Klapa, #02-01, Singapore 199329",
        iframe: "https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d3988.7846024345927!2d103.85963729999999!3d1.3042818999999999!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x31da1963987b59bf%3A0xc1ed652c0bc65836!2sUnsalted%20by%20Lazy%20Potato!5e0!3m2!1sen!2ssg!4v1759481675787!5m2!1sen!2ssg",
        website: "https://hitpay.shop/unsleeved/"
    }
];

export const MIN_SEARCH_LENGTH = 3;

export const API_BASE_URL = (window.location.hostname === "staging.gishathfetch.com" || window.location.hostname === "localhost")
    ? "https://staging-api.gishathfetch.com/"
    : "https://api.gishathfetch.com/";

export const BASE_URL = (window.location.hostname === "staging.gishathfetch.com" || window.location.hostname === "localhost")
    ? "https://staging.gishathfetch.com/"
    : "https://gishathfetch.com/";
