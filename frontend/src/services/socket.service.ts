const API_BASE_URL = "http://localhost"; // Adjust to Traefik address if different

export const getAuthHeader = () => {
    const token = localStorage.getItem("access_token");
    return token ? { Authorization: `Bearer ${token}` } : {};
};

export async function request<T>(
    path: string,
    options: RequestInit = {},
): Promise<T> {
    const url = `${API_BASE_URL}${path}`;
    const headers = {
        "Content-Type": "application/json",
        ...getAuthHeader(),
        ...options.headers,
    };

    const response = await fetch(url, { ...options, headers: headers as HeadersInit });

    if (!response.ok) {
        const error = await response
            .json()
            .catch(() => ({ error: "Unknown error" }));
        throw new Error(
            error.error ||
                error.message ||
                `Request failed with status ${response.status}`,
        );
    }

    if (response.status === 204) return {} as T;
    return response.json();
}
