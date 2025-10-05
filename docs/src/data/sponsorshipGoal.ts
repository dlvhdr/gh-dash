const GITHUB_TOKEN = import.meta.env.GITHUB_TOKEN;

export const getSponsorshipGoal = async () => {
  const query = `query {
    user(login:"dlvhdr") {
      monthlyEstimatedSponsorsIncomeInCents
    }
  }`;

  const response = await fetch("https://api.github.com/graphql", {
    method: "POST",
    body: JSON.stringify({ query }),
    headers: {
      Authorization: `bearer ${GITHUB_TOKEN}`,
    },
  }).then((r) => r.json());

  return response;
};
