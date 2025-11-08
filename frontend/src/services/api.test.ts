import { describe, it, expect, vi, beforeEach } from "vitest"
import {
  api,
  type Note,
  type CreateNoteInput,
  type UpdateNoteInput,
} from "./api"

globalThis.fetch = vi.fn()

describe("api", () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.stubEnv("VITE_API_BASE_URL", "http://localhost:4000/v1")
  })

  describe("createNote", () => {
    it("should create a note successfully", async () => {
      const mockNote: Note = {
        id: 1,
        title: "Test Note",
        body: "Test Body",
        tags: ["test"],
        archived: false,
        updated_at: "2024-01-01T00:00:00Z",
        version: 1,
      }

      ;(fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ note: mockNote }),
      })

      const input: CreateNoteInput = {
        title: "Test Note",
        body: "Test Body",
        tags: ["test"],
      }

      const result = await api.createNote(input)

      expect(result).toEqual(mockNote)
      expect(fetch).toHaveBeenCalledWith(
        "http://localhost:4000/v1/notes",
        expect.objectContaining({
          method: "POST",
          body: JSON.stringify(input),
          headers: expect.objectContaining({
            "Content-Type": "application/json",
          }),
        }),
      )
    })

    it("should throw error on failure", async () => {
      ;(fetch as any).mockResolvedValueOnce({
        ok: false,
        status: 400,
        json: async () => ({ error: "Bad request" }),
      })

      const input: CreateNoteInput = {
        title: "Test",
        body: "Body",
        tags: ["test"],
      }

      await expect(api.createNote(input)).rejects.toThrow("Bad request")
    })
  })

  describe("getNote", () => {
    it("should fetch a note successfully", async () => {
      const mockNote: Note = {
        id: 1,
        title: "Test Note",
        body: "Test Body",
        tags: ["test"],
        archived: false,
        updated_at: "2024-01-01T00:00:00Z",
        version: 1,
      }

      ;(fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ note: mockNote }),
      })

      const result = await api.getNote(1)

      expect(result).toEqual(mockNote)
      expect(fetch).toHaveBeenCalledWith(
        "http://localhost:4000/v1/notes/1",
        expect.objectContaining({
          headers: expect.objectContaining({
            "Content-Type": "application/json",
          }),
        }),
      )
    })
  })

  describe("updateNote", () => {
    it("should update a note successfully", async () => {
      const mockNote: Note = {
        id: 1,
        title: "Updated Note",
        body: "Updated Body",
        tags: ["updated"],
        archived: true,
        updated_at: "2024-01-01T00:00:00Z",
        version: 2,
      }

      ;(fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ note: mockNote }),
      })

      const input: UpdateNoteInput = {
        title: "Updated Note",
        body: "Updated Body",
        tags: ["updated"],
        archived: true,
      }

      const result = await api.updateNote(1, input)

      expect(result).toEqual(mockNote)
      expect(fetch).toHaveBeenCalledWith(
        "http://localhost:4000/v1/notes/1",
        expect.objectContaining({
          method: "PUT",
          body: JSON.stringify(input),
        }),
      )
    })
  })

  describe("deleteNote", () => {
    it("should delete a note successfully", async () => {
      ;(fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => ({}),
      })

      await api.deleteNote(1)

      expect(fetch).toHaveBeenCalledWith(
        "http://localhost:4000/v1/notes/1",
        expect.objectContaining({
          method: "DELETE",
        }),
      )
    })

    it("should throw error on failure", async () => {
      ;(fetch as any).mockResolvedValueOnce({
        ok: false,
        status: 404,
        json: async () => ({ error: "Not found" }),
      })

      await expect(api.deleteNote(999)).rejects.toThrow("Not found")
    })
  })
})
