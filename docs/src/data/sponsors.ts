const GH_TOKEN = import.meta.env.GH_TOKEN;

const EMPTY_RESPONSE = {};

type TiersResponse = Record<
  string,
  {
    tier: {
      id: string;
      name: string;
      monthlyPriceInDollars: number;
      title: string;
    };
    users: {
      login: string;
      avatarUrl: string;
      websiteUrl: string;
      url: string;
    }[];
  }
>;

export const getSponsorTiers = async (): Promise<TiersResponse> => {
  const tiersQuery = `query {
  user(login: "dlvhdr") {
    sponsorsListing {
      tiers(first: 20) {
        nodes {
          id
          monthlyPriceInDollars
          name
          isOneTime
          description
        }
      }
    }
  }
}`;

  const tiersResponse = await fetch("https://api.github.com/graphql", {
    method: "POST",
    body: JSON.stringify({ query: tiersQuery }),
    headers: {
      Authorization: `bearer ${GH_TOKEN}`,
    },
  });

  const tiersData = await tiersResponse.json();
  if (
    tiersResponse.status != 200 ||
    ("errors" in tiersData && tiersData.errors.length > 0)
  ) {
    console.error("[ERROR] failed fetching sponsorship tiers", {
      status: tiersResponse.status,
      json: JSON.stringify(tiersData, null, 2),
    });
    return EMPTY_RESPONSE;
  }

  const res: TiersResponse = {};

  const tiers: {
    id: string;
    monthlyPriceInDollars: number;
    name: string;
    isOneTime: boolean;
    description: string;
  }[] = tiersData.data.user.sponsorsListing.tiers.nodes;
  for (const tier of tiers) {
    if (tier.isOneTime) {
      continue;
    }
    const sponsorsQuery = `query{
  user(login: "dlvhdr") {
    sponsors(first: 100, tierId: "${tier.id}") {
      nodes {
        ... on User {
          login
          avatarUrl
          websiteUrl
          url
        }
      }
    }
  }
}`;

    const sponsorsResponse = await fetch("https://api.github.com/graphql", {
      method: "POST",
      body: JSON.stringify({ query: sponsorsQuery }),
      headers: {
        Authorization: `bearer ${GH_TOKEN}`,
      },
    });

    const sponsorData = await sponsorsResponse.json();
    if (sponsorsResponse.status === 200 && !("errors" in sponsorData)) {
      res[tier.name] = {
        tier: {
          ...tier,
          title: (
            tier.description.split("\n").at(0)?.replace("# ", "") ?? tier.name
          ).trim(),
        },
        users: sponsorData.data.user.sponsors.nodes.filter(
          (user: any) => "login" in user,
        ),
      };
    } else {
      console.error("[ERROR] failed fetching sponsors for tier", {
        status: tiersResponse.status,
        tier: tier.name,
        json: JSON.stringify(sponsorData, null, 2),
      });
    }
  }

  return res;
};
