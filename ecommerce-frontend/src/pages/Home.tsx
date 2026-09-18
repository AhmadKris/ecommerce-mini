import { Link } from "react-router-dom";

/** Landing page — entry point into the product catalog. */
export function Home() {
  return (
    <main className="mx-auto max-w-3xl px-6 py-16">
      <h1 className="text-display-lg text-(--ink-primary)">Belanja lebih tenang, dikelola lebih rapi.</h1>
      <p className="text-body-lg text-(--ink-secondary) mt-4 max-w-xl">
        Storefront dan admin panel dari satu design system yang sama — checkout menyusul di iterasi
        berikutnya.
      </p>
      <Link
        to="/products"
        className="text-label-md bg-primary-600 text-(--ink-inverse) hover:bg-primary-700 mt-8 inline-block rounded-md px-5 py-2.5 transition-colors"
      >
        Lihat produk
      </Link>
    </main>
  );
}
