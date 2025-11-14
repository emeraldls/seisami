import {
  GetCloudAPIURL,
  GetWebURL,
  GetCollabServerAddress,
} from "../../wailsjs/go/main/App";

const fixedCloudUrl = "https://cloud.seisami.hooklytics.com";
const fixedWebUrl = "https://seisami.hooklytics.com";
const fixedWebsocketUrl = "wss://cloud.seisami.hooklytics.com/ws";

const CLOUD_API_URL = import.meta.env.VITE_CLOUD_API_URL ?? fixedCloudUrl;
const WEB_URL = import.meta.env.VITE_WEB_URL ?? fixedWebUrl;
const WEBSOCKET_URL = import.meta.env.VITE_WEBSOCKET_URL ?? fixedWebsocketUrl;

export { CLOUD_API_URL, WEB_URL, WEBSOCKET_URL };

export async function loadConfig() {
  const [cloud, web, ws] = await Promise.all([
    GetCloudAPIURL(),
    GetWebURL(),
    GetCollabServerAddress(),
  ]);

  return {
    CLOUD_API_URL: cloud,
    WEB_URL: web,
    WEBSOCKET_URL: ws,
  };
}
