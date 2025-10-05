import type { APIRoute } from "astro";
const GH_TOKEN = import.meta.env.GITHUB_TOKEN;

export const GET: APIRoute = async () => {
  const query = `query {
    user(login:"dlvhdr") {
      monthlyEstimatedSponsorsIncomeInCents
    }
  }`;

  const response = await fetch("https://api.github.com/graphql", {
    method: "POST",
    body: JSON.stringify({ query }),
    headers: {
      Authorization: `bearer ${GH_TOKEN}`,
    },
  }).then((r) => r.json());

  return new Response(JSON.stringify(response), {
    status: 200,
    headers: {
      "Content-Type": "application/json",
    },
  });
};
