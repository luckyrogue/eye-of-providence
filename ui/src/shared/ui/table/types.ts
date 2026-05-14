export type DataTableLabels = {
  filterPlaceholder?: string;
  columnsButton?: string;
  toggleColumns?: string;
  noResults?: string;
  loading?: string;
  previous?: string;
  next?: string;
  firstPage?: string;
  lastPage?: string;
  rowsPerPage?: string;
  pageOfTotal?: (page: number, total: number) => string;
  selectedOfTotal?: (sel: number, total: number) => string;
};
