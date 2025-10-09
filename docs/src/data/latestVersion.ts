const GH_TOKEN = import.meta.env.GH_TOKEN;

export const getLatestVersion = async () => {
  const response = await fetch(
    "https://api.github.com/repos/dlvhdr/gh-dash/releases/latest",
    {
      method: "GET",
      headers: {
        Authorization: `bearer ${GH_TOKEN}`,
      },
    },
  );

  if (response.status != 200) {
    return { tag_name: "v4.0.0" };
  }

  const data: { tag_name: string } = await response.json();

  return data;
};
