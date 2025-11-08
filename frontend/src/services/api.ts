const API_BASE_URL = import.meta.env.VITE_API_BASE_URL

if (!API_BASE_URL) {
  throw new Error("VITE_API_BASE_URL environment variable is required")
}

export interface Note {
  id: number
  title: string
  body: string
  tags: string[]
  archived: boolean
  updated_at: string
  version: number
}

export interface CreateNoteInput {
  title: string
  body: string
  tags: string[]
}

export interface UpdateNoteInput {
  title: string
  body: string
  tags: string[]
  archived?: boolean
}

async function fetchAPI<T>(
  endpoint: string,
  options?: RequestInit,
): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...options?.headers,
    },
  })

  if (!response.ok) {
    let errorMessage = `HTTP error! status: ${response.status}`
    try {
      const errorData = (await response.json()) as { error?: string }
      errorMessage = errorData.error || errorMessage
    } catch {
      // Response body is not valid JSON, use default error message
    }
    throw new Error(errorMessage)
  }

  const data = (await response.json()) as T
  return data
}

export const api = {
  createNote: async (input: CreateNoteInput): Promise<Note> => {
    const data = await fetchAPI<{ note: Note }>("/notes", {
      method: "POST",
      body: JSON.stringify(input),
    })
    return data.note
  },

  getNote: async (id: number): Promise<Note> => {
    const data = await fetchAPI<{ note: Note }>(`/notes/${id}`)
    return data.note
  },

  updateNote: async (id: number, input: UpdateNoteInput): Promise<Note> => {
    const data = await fetchAPI<{ note: Note }>(`/notes/${id}`, {
      method: "PUT",
      body: JSON.stringify(input),
    })
    return data.note
  },

  deleteNote: async (id: number): Promise<void> => {
    await fetchAPI(`/notes/${id}`, {
      method: "DELETE",
    })
  },
}
