import { CatalogPlaceholder } from "@/components/catalog-placeholder";

export const metadata = { title: "Medical Materials | Qayen" };

export default function MedicalMaterialsPage() {
  return (
    <CatalogPlaceholder
      title="Medical Materials"
      description="Browse medical equipment and materials."
      icon="medical"
    />
  );
}
