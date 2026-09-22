const version = "v1";
// Use the same host the page was loaded from instead of a hardcoded
// "localhost" — "localhost" always means "this device", so if you open
// the site from another device (or via your machine's network IP) it
// would otherwise try to reach a backend on THAT device, not your server.
const getBackendHost = () => {
  if (typeof window !== "undefined") {
    return window.location.hostname;
  }
  return "localhost";
};

export const BASE_URL = `http://${getBackendHost()}:4000/api/${version}`;