import { defineStore } from 'pinia'
import { ref } from 'vue'
import { bookAPI } from '@/api'
import type { Book } from '@/types'

export const useBookStore = defineStore('book', () => {
  const books = ref<Book[]>([])
  const currentBook = ref<Book | null>(null)

  const fetchBooks = async () => {
    const res = await bookAPI.getAll()
    if (res.success && res.data) {
      books.value = res.data
    }
  }

  const fetchBook = async (id: number) => {
    const res = await bookAPI.getOne(id)
    if (res.success && res.data) {
      currentBook.value = res.data
    }
  }

  const createBook = async (data: Partial<Book>) => {
    const res = await bookAPI.create(data)
    if (res.success && res.data) {
      books.value.unshift(res.data)
      return res.data
    }
  }

  const updateBook = async (id: number, data: Partial<Book>) => {
    const res = await bookAPI.update(id, data)
    if (res.success && res.data) {
      const index = books.value.findIndex(b => b.id === id)
      if (index !== -1) {
        books.value[index] = res.data
      }
      if (currentBook.value?.id === id) {
        currentBook.value = res.data
      }
      return res.data
    }
  }

  const deleteBook = async (id: number) => {
    const res = await bookAPI.delete(id)
    if (res.success) {
      books.value = books.value.filter(b => b.id !== id)
      if (currentBook.value?.id === id) {
        currentBook.value = null
      }
    }
  }

  return {
    books,
    currentBook,
    fetchBooks,
    fetchBook,
    createBook,
    updateBook,
    deleteBook
  }
})

