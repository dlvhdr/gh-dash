const GH_TOKEN = import.meta.env.GH_TOKEN;

export const getSponsorshipGoal = async () => {
  const query = `query { user(login:\"dlvhdr\") { monthlyEstimatedSponsorsIncomeInCents } }`;

  const response = await fetch("https://api.github.com/graphql", {
    method: "POST",
    body: JSON.stringify({ query }),
    headers: {
      Authorization: `bearer ${GH_TOKEN}`,
    },
  }).then((r) => r.json());

  return response;
};
