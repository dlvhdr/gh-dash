import type { APIRoute } from "astro";
const GITHUB_TOKEN = import.meta.env.GITHUB_TOKEN;

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
      Authorization: `bearer ${GITHUB_TOKEN}`,
    },
  }).then((r) => r.json());
  /* prettier-ignore */ // [🪲 dlv]
  console.log(`${new Date().toISOString()}[🪲 dlv] response:`, response);

  return new Response(JSON.stringify(response), {
    status: 200,
    headers: {
      "Content-Type": "application/json",
    },
  });
};
