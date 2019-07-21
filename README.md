# icestats

Icestats generates geoip stats from your Icecast server. These can later be used for graphing, generate reports, etc. Example:


```json
{
  "Total": 400,
  "Countries": {
    "Bolivia": {
      "Total": 2,
      "ISO": "BO",
      "Continent": "South America",
      "Cities": {
        "La Paz": {
          "Total": 1,
          "Geohash": "6mpd1hppq5yc"
        }
      }
    },
    "Brazil": {
      "Total": 1,
      "ISO": "BR",
      "Continent": "South America",
      "Cities": {
        "Belo Horizonte": {
          "Total": 1,
          "Geohash": "7h2wz9g2xt0j"
        }
      }
    },
    "Costa Rica": {
      "Total": 1,
      "ISO": "CR",
      "Continent": "North America",
      "Cities": {
        "Heredia": {
          "Total": 1,
          "Geohash": "d1u0vu5u3d0n"
        }
      }
    },
    "Mexico": {
      "Total": 4,
      "ISO": "MX",
      "Continent": "North America",
      "Cities": {
        "Mexico City": {
          "Total": 1,
          "Geohash": "9g3w81ckrjmy"
        },
        "Torreon": {
          "Total": 2,
          "Geohash": "9fvubjh3bs72"
        },
        "Villaflores": {
          "Total": 1,
          "Geohash": "9fvt5br0fxz2"
        }
      }
    },
    "Peru": {
      "Total": 1,
      "ISO": "PE",
      "Continent": "South America",
      "Cities": {
        "Lima": {
          "Total": 1,
          "Geohash": "6mc5qz50ssk8"
        }
      }
    }
  }
}
```
