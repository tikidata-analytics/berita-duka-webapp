package views

import (
	"fmt"
	"html"
	"net/url"
	"strings"
	"time"

	"github.com/tikidata-analytics/berita-duka-webapp/internal/models"
)

// FormatIndoDate converts a time.Time into an Indonesian localized readable date.
func FormatIndoDate(t time.Time) string {
	months := []string{
		"", "Januari", "Februari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember",
	}
	days := []string{"Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"}

	weekday := days[t.Weekday()]
	return fmt.Sprintf("%s, %d %s %d", weekday, t.Day(), months[t.Month()], t.Year())
}

// FormatIndoDateTime converts time.Time into Date + Time with WIB timezone indicator.
func FormatIndoDateTime(t time.Time) string {
	dateStr := FormatIndoDate(t)
	return fmt.Sprintf("%s pukul %02d:%02d WIB", dateStr, t.Hour(), t.Minute())
}

// RenderMicrosite generates a standalone HTML microsite with OpenGraph WhatsApp cards.
func RenderMicrosite(ob *models.Obituary, baseURL, siteName string) string {
	micrositeURL := fmt.Sprintf("%s/m/%s", baseURL, ob.Slug)

	photoURL := ob.PhotoURL
	if photoURL == "" {
		photoURL = "https://images.unsplash.com/photo-1518495973542-4542c06a5843?auto=format&fit=crop&w=800&q=80"
	}

	ogTitle := fmt.Sprintf("Berita Lelayu / Duka Cita - %s", ob.FullName)
	if ob.CallName != "" {
		ogTitle = fmt.Sprintf("Berita Duka - %s (%s)", ob.FullName, ob.CallName)
	}

	funeralStr := FormatIndoDateTime(ob.FuneralTime)
	ogDesc := fmt.Sprintf("Telah berpulang %s (Usia %d thn). Pemakaman/Kremasi pada %s di %s.",
		ob.FullName, ob.Age, funeralStr, ob.CemeteryName)

	waShareText := url.QueryEscape(fmt.Sprintf(
		"Berita Duka Cita / Lelayu:\n\nTelah berpulang ke rahmatullah/pangkuan Bapa di Surga:\n* %s *\n(Usia %d tahun)\n\nJadwal Pemakaman: %s\nTempat: %s\n\nDetail lengkap jadwal & lokasi prosesi:\n%s",
		ob.FullName, ob.Age, funeralStr, ob.CemeteryName, micrositeURL,
	))

	// Generate Family Members HTML
	var familyHTML strings.Builder
	if len(ob.FamilyMembers) > 0 {
		familyHTML.WriteString(`<div class="space-y-4">`)
		for _, group := range ob.FamilyMembers {
			familyHTML.WriteString(fmt.Sprintf(`
				<div class="border-b border-stone-800/60 pb-3">
					<div class="text-xs uppercase tracking-widest text-amber-500/90 font-medium mb-1">%s</div>
					<div class="text-stone-200 text-sm leading-relaxed font-serif">%s</div>
				</div>
			`, html.EscapeString(group.Role), html.EscapeString(strings.Join(group.Names, " • "))))
		}
		familyHTML.WriteString(`</div>`)
	} else {
		familyHTML.WriteString(`<p class="text-stone-400 text-sm italic font-serif">Keluarga besar yang ditinggalkan.</p>`)
	}

	// Ceremony Time HTML
	var ceremonyHTML string
	if ob.CeremonyTime != nil {
		ceremonyHTML = fmt.Sprintf(`
			<div class="bg-stone-900/60 rounded-xl p-4 border border-stone-800/80">
				<div class="text-xs uppercase tracking-wider text-stone-400 font-semibold mb-1">Ibadah / Doa Pelepasan</div>
				<div class="text-stone-100 font-medium text-sm md:text-base">%s</div>
			</div>
		`, FormatIndoDateTime(*ob.CeremonyTime))
	}

	// Funeral Home HTML
	var funeralHomeHTML string
	if ob.FuneralHomeName != "" {
		roomStr := ""
		if ob.FuneralHomeRoom != "" {
			roomStr = fmt.Sprintf(" (%s)", html.EscapeString(ob.FuneralHomeRoom))
		}
		funeralHomeHTML = fmt.Sprintf(`
			<div class="bg-stone-900/60 rounded-xl p-4 border border-stone-800/80">
				<div class="text-xs uppercase tracking-wider text-stone-400 font-semibold mb-1">Disemayamkan di</div>
				<div class="text-stone-100 font-medium text-sm md:text-base">%s%s</div>
			</div>
		`, html.EscapeString(ob.FuneralHomeName), roomStr)
	}

	// Google Maps Button
	var mapsBtnHTML string
	if ob.CemeteryMapsURL != "" {
		mapsBtnHTML = fmt.Sprintf(`
			<a href="%s" target="_blank" rel="noopener noreferrer" 
			   class="inline-flex items-center justify-center w-full px-4 py-3 bg-amber-600 hover:bg-amber-500 text-white font-medium rounded-xl transition duration-200 text-sm shadow-md shadow-amber-950/40">
				📍 Buka Rute Navigasi Google Maps
			</a>
		`, html.EscapeString(ob.CemeteryMapsURL))
	}

	// Donation / Condolence Bank Account
	var bankHTML string
	if ob.BankName != "" && ob.BankAccountNumber != "" {
		bankHTML = fmt.Sprintf(`
			<div class="mt-8 p-5 bg-stone-900/80 rounded-2xl border border-stone-800 text-center space-y-2">
				<div class="text-xs uppercase tracking-wider text-amber-500 font-semibold">Tanda Kasih & Duka Cita (Digital)</div>
				<div class="text-xs text-stone-400">Bagi kerabat yang ingin menyampaikan tanda kasih untuk keluarga:</div>
				<div class="inline-block bg-stone-950 px-4 py-2 rounded-lg border border-stone-800/90 font-mono text-stone-100 text-sm">
					%s - <span class="font-bold text-amber-400">%s</span>
				</div>
				<div class="text-xs text-stone-300">a.n. %s</div>
			</div>
		`, html.EscapeString(ob.BankName), html.EscapeString(ob.BankAccountNumber), html.EscapeString(ob.BankAccountName))
	}

	birthStr := ""
	if ob.BirthDate != nil {
		birthStr = fmt.Sprintf("%s — ", FormatIndoDate(*ob.BirthDate))
	}
	lifespanStr := fmt.Sprintf("%s%s", birthStr, FormatIndoDate(ob.DeathDate))
	if ob.Age > 0 {
		lifespanStr += fmt.Sprintf(" (Usia %d Tahun)", ob.Age)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="id">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>%s | %s</title>
	
	<!-- OpenGraph & WhatsApp Social Preview -->
	<meta property="og:title" content="%s">
	<meta property="og:description" content="%s">
	<meta property="og:image" content="%s">
	<meta property="og:url" content="%s">
	<meta property="og:type" content="article">
	
	<meta name="twitter:card" content="summary_large_image">
	<meta name="twitter:title" content="%s">
	<meta name="twitter:description" content="%s">
	<meta name="twitter:image" content="%s">

	<!-- Modern Tailwind CSS via CDN -->
	<script src="https://cdn.tailwindcss.com"></script>
	<link rel="preconnect" href="https://fonts.googleapis.com">
	<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
	<link href="https://fonts.googleapis.com/css2?family=Cinzel:wght@500;700&family=Plus+Jakarta+Sans:wght@300;400;500;600&family=Playfair+Display:ital,wght@0,400;0,600;1,400&display=swap" rel="stylesheet">
	
	<style>
		body { font-family: 'Plus Jakarta Sans', sans-serif; }
		.font-serif-title { font-family: 'Cinzel', serif; }
		.font-serif { font-family: 'Playfair Display', serif; }
	</style>
</head>
<body class="bg-stone-950 text-stone-200 min-h-screen py-8 px-4 flex flex-col items-center">

	<main class="w-full max-w-xl bg-stone-900/90 rounded-3xl border border-stone-800 shadow-2xl overflow-hidden backdrop-blur-md">
		
		<!-- Header Banner & Photo -->
		<div class="relative pt-12 pb-6 px-6 text-center bg-gradient-to-b from-stone-900 to-stone-950 border-b border-stone-800/80">
			<div class="text-xs uppercase tracking-widest text-stone-400 font-serif-title mb-4">Mengingat & Menghormati</div>
			
			<div class="relative mx-auto w-36 h-36 md:w-44 md:h-44 rounded-full p-1.5 bg-gradient-to-tr from-amber-600/40 via-stone-700 to-amber-500/30 shadow-2xl mb-5">
				<img src="%s" alt="%s" class="w-full h-full object-cover rounded-full filter grayscale hover:grayscale-0 transition duration-500 border border-stone-800">
			</div>

			<div class="text-xs tracking-widest text-amber-500/80 uppercase font-medium mb-1">Berita Duka Cita</div>
			<h1 class="text-2xl md:text-3xl font-serif font-bold text-stone-100 tracking-wide mb-1">%s</h1>
			<p class="text-stone-400 text-xs md:text-sm font-serif italic mb-3">%s</p>
			<div class="inline-block px-4 py-1 rounded-full bg-stone-800/80 text-amber-400/90 text-xs font-mono border border-stone-700/60">
				%s
			</div>
		</div>

		<!-- Body Content -->
		<div class="p-6 md:p-8 space-y-8">
			
			<!-- Kenangan / Bio -->
			%s

			<!-- Jadwal & Lokasi Prosesi -->
			<section class="space-y-4">
				<div class="flex items-center space-x-2 border-b border-stone-800 pb-2">
					<span class="text-amber-500 font-bold">✦</span>
					<h2 class="text-xs uppercase tracking-widest font-serif-title text-stone-300 font-bold">Jadwal & Lokasi Prosesi</h2>
				</div>

				<div class="grid grid-cols-1 gap-3 text-stone-200">
					%s
					%s
					<div class="bg-stone-900/60 rounded-xl p-4 border border-stone-800/80">
						<div class="text-xs uppercase tracking-wider text-amber-400 font-semibold mb-1">Pemakaman / Kremasi</div>
						<div class="text-stone-100 font-semibold text-sm md:text-base mb-1">%s</div>
						<div class="text-stone-300 text-xs md:text-sm">Lokasi: <span class="text-stone-100 font-medium">%s</span></div>
					</div>
				</div>

				%s
			</section>

			<!-- Keluarga Berduka -->
			<section class="space-y-4">
				<div class="flex items-center space-x-2 border-b border-stone-800 pb-2">
					<span class="text-amber-500 font-bold">✦</span>
					<h2 class="text-xs uppercase tracking-widest font-serif-title text-stone-300 font-bold">Keluarga yang Mengasihi</h2>
				</div>
				%s
			</section>

			<!-- Tanda Kasih / Donasi Rekening -->
			%s

			<!-- Share & Action Buttons -->
			<div class="pt-6 border-t border-stone-800/80 space-y-3">
				<a href="https://api.whatsapp.com/send?text=%s" target="_blank" rel="noopener noreferrer"
				   class="flex items-center justify-center space-x-2 w-full py-3 bg-emerald-700 hover:bg-emerald-600 text-white rounded-xl font-medium text-sm transition duration-200 shadow-lg shadow-emerald-950/40">
					<span>📲 Bagikan ke Grup WhatsApp</span>
				</a>
				
				<div class="text-center">
					<a href="%s" class="text-xs text-stone-400 hover:text-stone-200 transition">
						← Kembali ke Halaman Utama
					</a>
				</div>
			</div>

		</div>

		<!-- Footer -->
		<footer class="bg-stone-950/80 py-4 px-6 border-t border-stone-800/80 text-center text-stone-400 text-xs">
			%s • Layanan Informasi Lelayu & Obituari Digital
		</footer>
	</main>

</body>
</html>`,
		html.EscapeString(ob.FullName), html.EscapeString(siteName),
		html.EscapeString(ogTitle), html.EscapeString(ogDesc), html.EscapeString(photoURL), html.EscapeString(micrositeURL),
		html.EscapeString(ogTitle), html.EscapeString(ogDesc), html.EscapeString(photoURL),
		html.EscapeString(photoURL), html.EscapeString(ob.FullName),
		html.EscapeString(ob.FullName),
		html.EscapeString(ob.CallName),
		html.EscapeString(lifespanStr),
		renderBio(ob.Bio),
		funeralHomeHTML, ceremonyHTML,
		html.EscapeString(funeralStr), html.EscapeString(ob.CemeteryName),
		mapsBtnHTML,
		familyHTML.String(),
		bankHTML,
		waShareText,
		baseURL,
		html.EscapeString(siteName),
	)
}

func renderBio(bio string) string {
	if strings.TrimSpace(bio) == "" {
		return ""
	}
	return fmt.Sprintf(`
		<div class="p-5 bg-stone-950/50 rounded-2xl border border-stone-800/70 text-center">
			<p class="font-serif italic text-stone-300 text-sm md:text-base leading-relaxed">
				"%s"
			</p>
		</div>
	`, html.EscapeString(bio))
}

// RenderCreateForm generates the responsive self-service submission form for family members with Cloudinary upload support.
func RenderCreateForm(baseURL, siteName string, cloudinaryEnabled bool) string {
	cldNotice := "Cloudinary aktif: Foto yang dipilih akan langsung diunggah otomatis."
	if !cloudinaryEnabled {
		cldNotice = "Cloudinary belum dikonfigurasi di server: Gunakan tautan URL foto langsung atau atur kredensial Cloudinary di .env."
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="id">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Buat Berita Duka Cita | %s</title>
	<script src="https://cdn.tailwindcss.com"></script>
	<link href="https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@300;400;500;600;700&family=Playfair+Display:wght@500;700&display=swap" rel="stylesheet">
	<style>
		body { font-family: 'Plus Jakarta Sans', sans-serif; }
		.font-serif { font-family: 'Playfair Display', serif; }
	</style>
</head>
<body class="bg-stone-950 text-stone-100 min-h-screen py-8 px-4 flex flex-col items-center">

	<main class="w-full max-w-2xl bg-stone-900 rounded-3xl border border-stone-800 p-6 md:p-10 shadow-2xl">
		<header class="text-center mb-8 border-b border-stone-800 pb-6">
			<a href="%s" class="text-xs uppercase tracking-widest text-amber-500 font-semibold">← Kembali ke Beranda</a>
			<h1 class="text-2xl md:text-3xl font-serif font-bold text-stone-100 mt-2">Buat Pengumuman Berita Duka</h1>
			<p class="text-stone-400 text-xs md:text-sm mt-1">Formulir mandiri keluarga untuk pembuatan microsite berita lelayu digital.</p>
		</header>

		<form id="obituaryForm" class="space-y-6">
			
			<!-- Bagian 1: Data Almarhum -->
			<div class="space-y-4">
				<h2 class="text-sm font-semibold tracking-wider uppercase text-amber-400 border-b border-stone-800 pb-1">1. Informasi Almarhum / Almarhumah</h2>
				
				<div>
					<label class="block text-xs font-medium text-stone-300 mb-1">Nama Lengkap & Gelar *</label>
					<input type="text" name="full_name" required placeholder="Contoh: Bpk. H. Bambang Soekoco, S.E." 
					       class="w-full px-4 py-2.5 bg-stone-950 border border-stone-800 rounded-xl text-stone-100 text-sm focus:outline-none focus:border-amber-500">
				</div>

				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div>
						<label class="block text-xs font-medium text-stone-300 mb-1">Nama Panggilan / Alias</label>
						<input type="text" name="call_name" placeholder="Contoh: Opa Bambang" 
						       class="w-full px-4 py-2.5 bg-stone-950 border border-stone-800 rounded-xl text-stone-100 text-sm focus:outline-none focus:border-amber-500">
					</div>
					<div>
						<label class="block text-xs font-medium text-stone-300 mb-1">Usia (Tahun)</label>
						<input type="number" name="age" placeholder="Contoh: 72" 
						       class="w-full px-4 py-2.5 bg-stone-950 border border-stone-800 rounded-xl text-stone-100 text-sm focus:outline-none focus:border-amber-500">
					</div>
				</div>

				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div>
						<label class="block text-xs font-medium text-stone-300 mb-1">Tanggal Lahir (Opsional)</label>
						<input type="date" name="birth_date" 
						       class="w-full px-4 py-2.5 bg-stone-950 border border-stone-800 rounded-xl text-stone-100 text-sm focus:outline-none focus:border-amber-500">
					</div>
					<div>
						<label class="block text-xs font-medium text-stone-300 mb-1">Tanggal Wafat *</label>
						<input type="date" name="death_date" required 
						       class="w-full px-4 py-2.5 bg-stone-950 border border-stone-800 rounded-xl text-stone-100 text-sm focus:outline-none focus:border-amber-500">
					</div>
				</div>

				<!-- Cloudinary Photo Uploader Component -->
				<div class="space-y-2 p-4 bg-stone-950/70 border border-stone-800 rounded-2xl">
					<label class="block text-xs font-medium text-amber-400">Foto Almarhum / Almarhumah</label>
					
					<div class="flex items-center gap-4">
						<div id="previewContainer" class="w-20 h-20 rounded-full border-2 border-stone-700 bg-stone-900 overflow-hidden flex items-center justify-center shrink-0">
							<img id="photoPreview" src="https://images.unsplash.com/photo-1518495973542-4542c06a5843?auto=format&fit=crop&w=200&q=80" 
							     alt="Preview" class="w-full h-full object-cover filter grayscale">
						</div>
						<div class="space-y-1 flex-1">
							<input type="file" id="photoFileInput" accept="image/*" class="block w-full text-xs text-stone-400
							       file:mr-4 file:py-2 file:px-4 file:rounded-xl file:border-0
							       file:text-xs file:font-medium file:bg-amber-600/80 file:text-white
							       hover:file:bg-amber-600 file:cursor-pointer">
							<div id="uploadStatus" class="text-xs text-stone-400 italic">Pilih foto dari galeri atau kamera ponsel</div>
						</div>
					</div>

					<input type="hidden" name="photo_url" id="photoUrlInput" value="">

					<div class="pt-2 border-t border-stone-800/80">
						<details class="text-xs text-stone-400 cursor-pointer">
							<summary class="hover:text-stone-300">Atau masukkan link gambar manual (URL)</summary>
							<div class="mt-2">
								<input type="url" id="manualUrlInput" placeholder="https://res.cloudinary.com/..." 
								       class="w-full px-3 py-2 bg-stone-900 border border-stone-800 rounded-xl text-stone-200 text-xs focus:outline-none focus:border-amber-500">
							</div>
						</details>
					</div>
					<p class="text-[11px] text-stone-400 italic">%s</p>
				</div>

				<div>
					<label class="block text-xs font-medium text-stone-300 mb-1">Sepatah Kata / Kenangan Singkat</label>
					<textarea name="bio" rows="2" placeholder="Contoh: Telah berpulang dengan tenang meninggalkan teladan kasih bagi seluruh anak dan cucu..." 
					          class="w-full px-4 py-2 bg-stone-950 border border-stone-800 rounded-xl text-stone-100 text-sm focus:outline-none focus:border-amber-500"></textarea>
				</div>
			</div>

			<!-- Bagian 2: Jadwal & Lokasi -->
			<div class="space-y-4 pt-4 border-t border-stone-800">
				<h2 class="text-sm font-semibold tracking-wider uppercase text-amber-400 border-b border-stone-800 pb-1">2. Jadwal & Lokasi Prosesi</h2>

				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div>
						<label class="block text-xs font-medium text-stone-300 mb-1">Rumah Duka (Jika Disemayamkan)</label>
						<input type="text" name="funeral_home_name" placeholder="Contoh: Rumah Duka Dharmais" 
						       class="w-full px-4 py-2.5 bg-stone-950 border border-stone-800 rounded-xl text-stone-100 text-sm focus:outline-none focus:border-amber-500">
					</div>
					<div>
						<label class="block text-xs font-medium text-stone-300 mb-1">Nama Ruangan / Kamar</label>
						<input type="text" name="funeral_home_room" placeholder="Contoh: Ruang VIP C" 
						       class="w-full px-4 py-2.5 bg-stone-950 border border-stone-800 rounded-xl text-stone-100 text-sm focus:outline-none focus:border-amber-500">
					</div>
				</div>

				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div>
						<label class="block text-xs font-medium text-stone-300 mb-1">Waktu Ibadah / Doa Pelepasan</label>
						<input type="datetime-local" name="ceremony_time" 
						       class="w-full px-4 py-2.5 bg-stone-950 border border-stone-800 rounded-xl text-stone-100 text-sm focus:outline-none focus:border-amber-500">
					</div>
					<div>
						<label class="block text-xs font-medium text-stone-300 mb-1">Waktu Pemakaman / Kremasi *</label>
						<input type="datetime-local" name="funeral_time" required 
						       class="w-full px-4 py-2.5 bg-stone-950 border border-stone-800 rounded-xl text-stone-100 text-sm focus:outline-none focus:border-amber-500">
					</div>
				</div>

				<div>
					<label class="block text-xs font-medium text-stone-300 mb-1">Nama Tempat Pemakaman / Krematorium *</label>
					<input type="text" name="cemetery_name" required placeholder="Contoh: TPU Karet Bivak / Krematorium Oasis" 
					       class="w-full px-4 py-2.5 bg-stone-950 border border-stone-800 rounded-xl text-stone-100 text-sm focus:outline-none focus:border-amber-500">
				</div>

				<div>
					<label class="block text-xs font-medium text-stone-300 mb-1">Tautan Google Maps Lokasi</label>
					<input type="url" name="cemetery_maps_url" placeholder="https://maps.app.goo.gl/..." 
					       class="w-full px-4 py-2.5 bg-stone-950 border border-stone-800 rounded-xl text-stone-100 text-sm focus:outline-none focus:border-amber-500">
				</div>
			</div>

			<!-- Bagian 3: Susunan Keluarga Berduka -->
			<div class="space-y-4 pt-4 border-t border-stone-800">
				<h2 class="text-sm font-semibold tracking-wider uppercase text-amber-400 border-b border-stone-800 pb-1">3. Susunan Keluarga yang Ditinggalkan</h2>
				
				<div class="space-y-2">
					<label class="block text-xs font-medium text-stone-300">Daftar Keluarga (Format: Satu peran per baris, pisahkan nama dengan koma)</label>
					<textarea id="family_input" rows="4" placeholder="Istri: Ny. Hj. Siti Aminah&#10;Anak & Menantu: Budi & Rina, Hendra & Dian&#10;Cucu: Kevin, Jessica, Sarah" 
					          class="w-full px-4 py-2 bg-stone-950 border border-stone-800 rounded-xl text-stone-100 text-sm font-mono focus:outline-none focus:border-amber-500"></textarea>
				</div>
			</div>

			<!-- Bagian 4: Rekening Tanda Kasih (Opsional) -->
			<div class="space-y-4 pt-4 border-t border-stone-800">
				<h2 class="text-sm font-semibold tracking-wider uppercase text-amber-400 border-b border-stone-800 pb-1">4. Rekening Duka Cita (Opsional)</h2>
				
				<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
					<div>
						<label class="block text-xs font-medium text-stone-300 mb-1">Nama Bank</label>
						<input type="text" name="bank_name" placeholder="BCA / Mandiri" 
						       class="w-full px-4 py-2.5 bg-stone-950 border border-stone-800 rounded-xl text-stone-100 text-sm focus:outline-none focus:border-amber-500">
					</div>
					<div>
						<label class="block text-xs font-medium text-stone-300 mb-1">Nomor Rekening</label>
						<input type="text" name="bank_account_number" placeholder="1234567890" 
						       class="w-full px-4 py-2.5 bg-stone-950 border border-stone-800 rounded-xl text-stone-100 text-sm focus:outline-none focus:border-amber-500">
					</div>
					<div>
						<label class="block text-xs font-medium text-stone-300 mb-1">Atas Nama</label>
						<input type="text" name="bank_account_name" placeholder="Budi Santoso" 
						       class="w-full px-4 py-2.5 bg-stone-950 border border-stone-800 rounded-xl text-stone-100 text-sm focus:outline-none focus:border-amber-500">
					</div>
				</div>
			</div>

			<div class="pt-6 border-t border-stone-800">
				<button type="submit" id="submitBtn" 
				        class="w-full py-3.5 bg-amber-600 hover:bg-amber-500 text-white font-semibold rounded-xl transition duration-200 shadow-lg shadow-amber-950/40">
					Terbitkan Berita Duka & Buat Microsite ➔
				</button>
			</div>
		</form>

		<div id="resultModal" class="hidden fixed inset-0 bg-black/80 flex items-center justify-center p-4 backdrop-blur-sm z-50">
			<div class="bg-stone-900 border border-stone-800 p-6 rounded-2xl max-w-md w-full text-center space-y-4">
				<div class="text-4xl">🕊️</div>
				<h3 class="text-lg font-bold text-stone-100">Microsite Berhasil Dibuat!</h3>
				<p class="text-stone-300 text-xs leading-relaxed" id="resultDesc">Pengumuman lelayu siap dibagikan kepada kerabat dan keluarga.</p>
				<div class="pt-2">
					<a id="micrositeLink" href="#" target="_blank" 
					   class="inline-block w-full py-2.5 bg-amber-600 hover:bg-amber-500 text-white font-medium rounded-xl text-sm transition">
						Buka Microsite Sekarang ➔
					</a>
				</div>
			</div>
		</div>
	</main>

	<script>
		const form = document.getElementById('obituaryForm');
		const submitBtn = document.getElementById('submitBtn');
		const resultModal = document.getElementById('resultModal');
		const micrositeLink = document.getElementById('micrositeLink');
		const photoFileInput = document.getElementById('photoFileInput');
		const photoPreview = document.getElementById('photoPreview');
		const photoUrlInput = document.getElementById('photoUrlInput');
		const manualUrlInput = document.getElementById('manualUrlInput');
		const uploadStatus = document.getElementById('uploadStatus');

		// Handle asynchronous Cloudinary file upload
		photoFileInput.addEventListener('change', async (e) => {
			const file = e.target.files[0];
			if (!file) return;

			// Preview locally immediately
			photoPreview.src = URL.createObjectURL(file);
			photoPreview.classList.remove('grayscale');
			uploadStatus.innerText = '⏳ Mengunggah foto ke Cloudinary...';
			uploadStatus.className = 'text-xs text-amber-400 font-medium';

			const formData = new FormData();
			formData.append('photo', file);

			try {
				const res = await fetch('%s/api/upload', {
					method: 'POST',
					body: formData
				});
				const data = await res.json();
				if (!res.ok) throw new Error(data.error || 'Gagal mengunggah foto');

				photoUrlInput.value = data.url;
				manualUrlInput.value = data.url;
				photoPreview.src = data.url;
				uploadStatus.innerText = '✓ Foto berhasil tersimpan di Cloudinary';
				uploadStatus.className = 'text-xs text-emerald-400 font-medium';
			} catch (err) {
				uploadStatus.innerText = '⚠️ ' + err.message;
				uploadStatus.className = 'text-xs text-rose-400';
			}
		});

		manualUrlInput.addEventListener('input', (e) => {
			const val = e.target.value.trim();
			photoUrlInput.value = val;
			if (val) {
				photoPreview.src = val;
				photoPreview.classList.remove('grayscale');
			}
		});

		form.addEventListener('submit', async (e) => {
			e.preventDefault();
			submitBtn.disabled = true;
			submitBtn.innerText = 'Menerbitkan...';

			const formData = new FormData(form);
			const data = Object.fromEntries(formData.entries());

			// Parse Family input
			const familyRaw = document.getElementById('family_input').value;
			const familyMembers = [];
			if (familyRaw.trim()) {
				const lines = familyRaw.split('\n');
				for (const line of lines) {
					const parts = line.split(':');
					if (parts.length >= 2) {
						const role = parts[0].trim();
						const names = parts.slice(1).join(':').split(',').map(s => s.trim()).filter(Boolean);
						if (role && names.length > 0) {
							familyMembers.push({ role, names });
						}
					}
				}
			}
			data.family_members = familyMembers;
			data.age = parseInt(data.age, 10) || 0;

			try {
				const res = await fetch('%s/api/obituaries', {
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify(data)
				});
				const result = await res.json();
				if (!res.ok) throw new Error(result.error || 'Gagal menerbitkan pengumuman');
				
				micrositeLink.href = result.microsite_url;
				resultModal.classList.remove('hidden');
			} catch (err) {
				alert('Error: ' + err.message);
			} finally {
				submitBtn.disabled = false;
				submitBtn.innerText = 'Terbitkan Berita Duka & Buat Microsite ➔';
			}
		});
	</script>

</body>
</html>`,
		html.EscapeString(siteName), baseURL, html.EscapeString(cldNotice), baseURL, baseURL,
	)
}

// RenderHomePage generates the portal home page showing recent announcements.
func RenderHomePage(obituaries []models.Obituary, baseURL, siteName string) string {
	var cardsHTML strings.Builder
	if len(obituaries) == 0 {
		cardsHTML.WriteString(`
			<div class="col-span-full py-16 text-center text-stone-400 font-serif italic">
				Belum ada pengumuman duka cita aktif.
			</div>
		`)
	} else {
		for _, ob := range obituaries {
			photoURL := ob.PhotoURL
			if photoURL == "" {
				photoURL = "https://images.unsplash.com/photo-1518495973542-4542c06a5843?auto=format&fit=crop&w=400&q=80"
			}
			funeralDateStr := FormatIndoDate(ob.FuneralTime)
			cardsHTML.WriteString(fmt.Sprintf(`
				<a href="%s/m/%s" class="group bg-stone-900 border border-stone-800 rounded-2xl overflow-hidden hover:border-amber-600/60 transition duration-300 flex flex-col shadow-lg">
					<div class="h-44 overflow-hidden relative bg-stone-950">
						<img src="%s" alt="%s" class="w-full h-full object-cover group-hover:scale-105 transition duration-500 filter grayscale group-hover:grayscale-0">
						<div class="absolute inset-0 bg-gradient-to-t from-stone-900 via-transparent"></div>
						<div class="absolute bottom-2 left-3 text-xs text-amber-400/90 font-mono">
							Wafat: %s
						</div>
					</div>
					<div class="p-4 flex-1 flex flex-col justify-between">
						<div>
							<h3 class="text-base font-serif font-bold text-stone-100 group-hover:text-amber-400 transition">%s</h3>
							<p class="text-xs text-stone-400 mt-1">Pemakaman: %s (%s)</p>
						</div>
						<div class="mt-4 pt-3 border-t border-stone-800/80 flex items-center justify-between text-xs text-amber-500 font-medium">
							<span>Lihat Detail & Jadwal</span>
							<span>➔</span>
						</div>
					</div>
				</a>
			`, baseURL, ob.Slug, html.EscapeString(photoURL), html.EscapeString(ob.FullName),
				FormatIndoDate(ob.DeathDate), html.EscapeString(ob.FullName),
				funeralDateStr, html.EscapeString(ob.CemeteryName)))
		}
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="id">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>%s</title>
	<script src="https://cdn.tailwindcss.com"></script>
	<link href="https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@300;400;500;600;700&family=Playfair+Display:wght@500;700&display=swap" rel="stylesheet">
	<style>
		body { font-family: 'Plus Jakarta Sans', sans-serif; }
		.font-serif { font-family: 'Playfair Display', serif; }
	</style>
</head>
<body class="bg-stone-950 text-stone-100 min-h-screen py-8 px-4 flex flex-col items-center">

	<div class="w-full max-w-5xl space-y-10">
		<header class="flex flex-col md:flex-row items-center justify-between gap-4 border-b border-stone-800 pb-6">
			<div>
				<div class="text-xs uppercase tracking-widest text-amber-500 font-semibold">Portal Informasi Duka Cita</div>
				<h1 class="text-2xl md:text-3xl font-serif font-bold text-stone-100 mt-1">%s</h1>
			</div>
			<div>
				<a href="%s/create" class="px-5 py-2.5 bg-amber-600 hover:bg-amber-500 text-white font-medium rounded-xl text-sm transition shadow-lg shadow-amber-950/40">
					+ Buat Pengumuman Duka Mandiri
				</a>
			</div>
		</header>

		<section class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-6">
			%s
		</section>

		<footer class="pt-8 border-t border-stone-800 text-center text-xs text-stone-400">
			%s • Powered by Go & Neon PostgreSQL (Deployable to Vercel)
		</footer>
	</div>

</body>
</html>`,
		html.EscapeString(siteName), html.EscapeString(siteName), baseURL,
		cardsHTML.String(), html.EscapeString(siteName),
	)
}
