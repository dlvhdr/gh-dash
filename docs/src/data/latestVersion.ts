const GH_TOKEN = import.meta.env.GH_TOKEN;

const repoUrlRegex = /https:\/\/github.com\/(.*)\/(.*)/;
const FALLBACK_TAGNAME = { tag_name: "v4.0.0" };

export const getLatestVersion = async (githubUrl: string) => {
  const matches = repoUrlRegex.exec(githubUrl);
  if (!matches || matches.length < 3) {
    return FALLBACK_TAGNAME;
  }
  const response = await fetch(
    `https://api.github.com/repos/${matches[1]}/${matches[2]}/releases/latest`,
    {
      method: "GET",
      headers: {
        Authorization: `bearer ${GH_TOKEN}`,
      },
    },
  );

  if (response.status != 200) {
    return FALLBACK_TAGNAME;
  }

  const data: { tag_name: string } = await response.json();

  return data;
};
