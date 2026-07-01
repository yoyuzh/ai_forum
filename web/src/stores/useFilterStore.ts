import { create } from "zustand";
import { FeedTab } from "../api/types";

interface FilterStore {
  feedTab: FeedTab;
  selectedCategory: string | null;
  searchQuery: string;
  selectedTags: string[];
  setFeedTab: (tab: FeedTab) => void;
  setCategory: (category: string | null) => void;
  setSearchQuery: (query: string) => void;
  toggleTag: (tag: string) => void;
  resetFilters: () => void;
}

export const useFilterStore = create<FilterStore>((set) => ({
  feedTab: "latest",
  selectedCategory: null,
  searchQuery: "",
  selectedTags: [],
  setFeedTab: (feedTab) => set({ feedTab }),
  setCategory: (selectedCategory) => set({ selectedCategory }),
  setSearchQuery: (searchQuery) => set({ searchQuery }),
  toggleTag: (tag) =>
    set((state) => {
      const active = state.selectedTags.includes(tag);
      return {
        selectedTags: active
          ? state.selectedTags.filter((t) => t !== tag)
          : [...state.selectedTags, tag],
      };
    }),
  resetFilters: () =>
    set({ feedTab: "latest", selectedCategory: null, searchQuery: "", selectedTags: [] }),
}));
