import axios from "axios";

const baseURL =
  import.meta.env.MODE === "development"
    ? "/api"
    : `https://${import.meta.env.VITE_API_URL}`;

export const api = axios.create({
  baseURL: baseURL,
  withCredentials: true,
});

api.interceptors.response.use(
  (res) => res,
  async (error) => {
    const originalRequest = error.config;
    if (
      error.response?.data.error.message === "Unauthorized" &&
      !originalRequest._retry
    ) {
      originalRequest._retry = true;
      const { data } = await api.get("/v1/refresh-tokens");

      api.defaults.headers.common["Authorization"] =
        `Bearer ${data.data.access_token}`;
      originalRequest.headers["Authorization"] =
        `Bearer ${data.data.access_token}`;
      return api(originalRequest);
    }
    return Promise.reject(error);
  },
);
