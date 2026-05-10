import type { TFunction } from "i18next";
import type { DataTableLabels, SelectColumnLabels } from "@eop/ui";

// Фабрика DataTable labels на основе common:data_table.
// Используем в каждой таблице, чтобы переводы шли через i18n, а не дублировались.
export function dtLabels(t: TFunction): DataTableLabels {
  return {
    filterPlaceholder: t("common:data_table.filter"),
    columnsButton: t("common:data_table.columns"),
    toggleColumns: t("common:data_table.toggle_columns"),
    noResults: t("common:data_table.no_results"),
    loading: t("common:data_table.loading"),
    previous: t("common:data_table.previous"),
    next: t("common:data_table.next"),
    firstPage: t("common:data_table.first_page"),
    lastPage: t("common:data_table.last_page"),
    rowsPerPage: t("common:data_table.rows_per_page"),
    pageOfTotal: (page, total) => t("common:data_table.page_of_total", { page, total }),
    selectedOfTotal: (selected, total) =>
      t("common:data_table.selected_of_total", { selected, total }),
  };
}

export function dtSelectLabels(t: TFunction): SelectColumnLabels {
  return {
    selectAll: t("common:data_table.select_all"),
    selectRow: t("common:data_table.select_row"),
  };
}
