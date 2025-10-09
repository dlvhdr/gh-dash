const GH_TOKEN = import.meta.env.GH_TOKEN;

export const getSponsorshipGoal = async () => {
  const query = `query { user(login:\"dlvhdr\") { monthlyEstimatedSponsorsIncomeInCents } }`;

  const response = await fetch("https://api.github.com/graphql", {
    method: "POST",
    body: JSON.stringify({ query }),
    headers: {
      Authorization: `bearer ${GH_TOKEN}`,
    },
  });

  if (response.status != 200) {
    return { data: { user: { monthlyEstimatedSponsorsIncomeInCents: 4000 } } };
  }

  const data: {
    data: { user: { monthlyEstimatedSponsorsIncomeInCents: number } };
  } = await response.json();

  return data;
};
