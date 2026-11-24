package assets

import (
	"bytes"
	"embed"
	"image"
	"image/color"
	_ "image/png"
	"path"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed sprites/characters/alfredo_*.png sprites/characters/girl_*.png sprites/characters/vip_aura.png
//go:embed sprites/world/boss_*.png sprites/world/bullet_*.png sprites/world/bus_*.png sprites/world/police_*.png sprites/world/shop_*.png sprites/world/stop_*.png sprites/world/tile_*.png
//go:embed sprites/map/building_*.png sprites/map/prop_*.png sprites/map/car_*.png sprites/map/tile_*.png
//go:embed sprites/ui/icon_*.png
var spriteFS embed.FS

// Expected sprite path constants matching the Sprite Bible and art subagent outputs.
const (
	// Alfredo (protagonist)
	SpriteAlfredoIdle0 = "sprites/characters/alfredo_idle_0.png"
	SpriteAlfredoIdle1 = "sprites/characters/alfredo_idle_1.png"
	SpriteAlfredoRun0  = "sprites/characters/alfredo_run_0.png"
	SpriteAlfredoRun1  = "sprites/characters/alfredo_run_1.png"
	SpriteAlfredoRun2  = "sprites/characters/alfredo_run_2.png"
	SpriteAlfredoRun3  = "sprites/characters/alfredo_run_3.png"
	SpriteAlfredoDash  = "sprites/characters/alfredo_dash.png"
	SpriteAlfredoThrow = "sprites/characters/alfredo_throw.png"

	// Alfredo skin: Neon Drip
	SpriteAlfredoDripIdle0 = "sprites/characters/alfredo_drip_idle_0.png"
	SpriteAlfredoDripIdle1 = "sprites/characters/alfredo_drip_idle_1.png"
	SpriteAlfredoDripRun0  = "sprites/characters/alfredo_drip_run_0.png"
	SpriteAlfredoDripRun1  = "sprites/characters/alfredo_drip_run_1.png"
	SpriteAlfredoDripRun2  = "sprites/characters/alfredo_drip_run_2.png"
	SpriteAlfredoDripRun3  = "sprites/characters/alfredo_drip_run_3.png"
	SpriteAlfredoDripDash  = "sprites/characters/alfredo_drip_dash.png"
	SpriteAlfredoDripThrow = "sprites/characters/alfredo_drip_throw.png"

	// Alfredo skin: Alfredo Banana
	SpriteAlfredoBananaIdle0 = "sprites/characters/alfredo_banana_idle_0.png"
	SpriteAlfredoBananaIdle1 = "sprites/characters/alfredo_banana_idle_1.png"
	SpriteAlfredoBananaRun0  = "sprites/characters/alfredo_banana_run_0.png"
	SpriteAlfredoBananaRun1  = "sprites/characters/alfredo_banana_run_1.png"
	SpriteAlfredoBananaRun2  = "sprites/characters/alfredo_banana_run_2.png"
	SpriteAlfredoBananaRun3  = "sprites/characters/alfredo_banana_run_3.png"
	SpriteAlfredoBananaDash  = "sprites/characters/alfredo_banana_dash.png"
	SpriteAlfredoBananaThrow = "sprites/characters/alfredo_banana_throw.png"

	// Alfredo skin: Aletta di Pollo
	SpriteAlfredoChickenIdle0 = "sprites/characters/alfredo_chicken_idle_0.png"
	SpriteAlfredoChickenIdle1 = "sprites/characters/alfredo_chicken_idle_1.png"
	SpriteAlfredoChickenRun0  = "sprites/characters/alfredo_chicken_run_0.png"
	SpriteAlfredoChickenRun1  = "sprites/characters/alfredo_chicken_run_1.png"
	SpriteAlfredoChickenRun2  = "sprites/characters/alfredo_chicken_run_2.png"
	SpriteAlfredoChickenRun3  = "sprites/characters/alfredo_chicken_run_3.png"
	SpriteAlfredoChickenDash  = "sprites/characters/alfredo_chicken_dash.png"
	SpriteAlfredoChickenThrow = "sprites/characters/alfredo_chicken_throw.png"

	// Alfredo skin: Cyber-Samurai Zero
	SpriteAlfredoSamuraiIdle0 = "sprites/characters/alfredo_samurai_idle_0.png"
	SpriteAlfredoSamuraiIdle1 = "sprites/characters/alfredo_samurai_idle_1.png"
	SpriteAlfredoSamuraiRun0  = "sprites/characters/alfredo_samurai_run_0.png"
	SpriteAlfredoSamuraiRun1  = "sprites/characters/alfredo_samurai_run_1.png"
	SpriteAlfredoSamuraiRun2  = "sprites/characters/alfredo_samurai_run_2.png"
	SpriteAlfredoSamuraiRun3  = "sprites/characters/alfredo_samurai_run_3.png"
	SpriteAlfredoSamuraiDash  = "sprites/characters/alfredo_samurai_dash.png"
	SpriteAlfredoSamuraiThrow = "sprites/characters/alfredo_samurai_throw.png"

	// Girls & VIP aura
	SpriteGirlKawaii0      = "sprites/characters/girl_kawaii_0.png"
	SpriteGirlKawaiiSheet  = "sprites/characters/girl_kawaii_sheet.png"
	SpriteGirlEmo0         = "sprites/characters/girl_emo_0.png"
	SpriteGirlEmoSheet     = "sprites/characters/girl_emo_sheet.png"
	SpriteGirlStudent0     = "sprites/characters/girl_student_0.png"
	SpriteGirlStudentSheet = "sprites/characters/girl_student_sheet.png"
	SpriteGirlGyaru0       = "sprites/characters/girl_gyaru_0.png"
	SpriteGirlGyaruSheet   = "sprites/characters/girl_gyaru_sheet.png"
	SpriteVIPAura          = "sprites/characters/vip_aura.png"

	// Bus stops & shop
	SpriteStopKawaii  = "sprites/world/stop_kawaii.png"
	SpriteStopEmo     = "sprites/world/stop_emo.png"
	SpriteStopStudent = "sprites/world/stop_student.png"
	SpriteStopGyaru   = "sprites/world/stop_gyaru.png"
	SpriteShopVending = "sprites/world/shop_vending.png"

	// Bus & police
	SpriteBusNeon       = "sprites/world/bus_neon.png"
	SpriteBusGold       = "sprites/world/bus_gold.png"
	SpritePoliceOfficer = "sprites/world/police_officer.png"
	SpritePoliceCar     = "sprites/world/police_car.png"

	// Bosses
	SpriteBossBestFriend = "sprites/world/boss_best_friend.png"
	SpriteBossLover      = "sprites/world/boss_lover.png"
	SpriteBossProfessor  = "sprites/world/boss_professor.png"
	SpriteBossExecutive  = "sprites/world/boss_executive.png"
	SpriteBossRector     = "sprites/world/boss_rector.png"
	SpriteBossMom        = "sprites/world/boss_mom.png"

	// Bullets
	SpriteBulletRazor      = "sprites/world/bullet_razor.png"
	SpriteBulletProtractor = "sprites/world/bullet_protractor.png"
	SpriteBulletTear       = "sprites/world/bullet_tear.png"
	SpriteBulletFlash      = "sprites/world/bullet_flash.png"
	SpriteBulletPolice     = "sprites/world/bullet_police.png"
	SpriteBulletCiabatta   = "sprites/world/bullet_ciabatta.png"
	SpriteBulletLipstick   = "sprites/world/bullet_lipstick.png"
	SpriteBulletContract   = "sprites/world/bullet_contract.png"
	SpriteBulletSeal       = "sprites/world/bullet_seal.png"

	// Arena tiles
	SpriteTileAsphalt   = "sprites/world/tile_asphalt.png"
	SpriteTileSidewalk  = "sprites/world/tile_sidewalk.png"
	SpriteTileZTL       = "sprites/world/tile_ztl.png"
	SpriteTileCrosswalk = "sprites/world/tile_crosswalk.png"

	// Layered Map Assets (buildings, landmarks, strategic props, cars, zone tiles)
	SpriteBuildingKonbini    = "sprites/map/building_konbini.png"
	SpriteBuildingCoffee     = "sprites/map/building_coffee.png"
	SpriteBuildingArcade     = "sprites/map/building_arcade.png"
	SpriteBuildingRamen      = "sprites/map/building_ramen.png"
	SpriteBuildingBookstore  = "sprites/map/building_bookstore.png"
	SpriteBuildingMetro      = "sprites/map/building_metro.png"
	SpriteBuildingNewsstand  = "sprites/map/building_newsstand.png"
	SpriteBuildingPizzeria   = "sprites/map/building_pizzeria.png"
	SpriteBuildingVinylBar   = "sprites/map/building_vinyl_bar.png"
	SpriteBuildingGelateria  = "sprites/map/building_gelateria.png"
	SpritePropFountain       = "sprites/map/prop_fountain.png"
	SpritePropArchGate       = "sprites/map/prop_arch_gate.png"
	SpritePropTreeLush       = "sprites/map/prop_tree_lush.png"
	SpritePropStreetlamp     = "sprites/map/prop_streetlamp.png"
	SpritePropBenchPlants    = "sprites/map/prop_bench_plants.png"
	SpritePropCatsClutter    = "sprites/map/prop_cats_clutter.png"
	SpritePropScooter        = "sprites/map/prop_scooter.png"
	SpritePropBikeRack       = "sprites/map/prop_bike_rack.png"
	SpritePropHydrantBin     = "sprites/map/prop_hydrant_bin.png"
	SpritePropZTLSign        = "sprites/map/prop_ztl_sign.png"
	SpritePropCanalBoat      = "sprites/map/prop_canal_boat.png"
	SpritePropSlotMachine    = "sprites/map/prop_slot_machine.png"
	SpritePropManhole        = "sprites/map/prop_manhole.png"
	SpritePropAsphaltCrack   = "sprites/map/prop_asphalt_crack.png"
	SpritePropFlowerBox      = "sprites/map/prop_flower_box.png"
	SpritePropNeonBillboard  = "sprites/map/prop_neon_billboard.png"
	SpritePropPhoneBooth     = "sprites/map/prop_phone_booth.png"
	SpritePropFoodCart       = "sprites/map/prop_food_cart.png"
	SpritePropKioskUmbrella  = "sprites/map/prop_kiosk_umbrella.png"
	SpritePropTrafficLight   = "sprites/map/prop_traffic_light.png"
	SpritePropBollardGlow    = "sprites/map/prop_bollard_glow.png"
	SpritePropPlanterNeon    = "sprites/map/prop_planter_neon.png"
	SpritePropVendingDrink   = "sprites/map/prop_vending_drink.png"
	SpritePropBusShelterDeco = "sprites/map/prop_bus_shelter_deco.png"
	SpriteCarRed             = "sprites/map/car_red.png"
	SpriteCarYellow          = "sprites/map/car_yellow.png"
	SpriteCarMint            = "sprites/map/car_mint.png"
	SpriteCarVanCyber        = "sprites/map/car_van_cyber.png"
	SpriteTileGrassPark      = "sprites/map/tile_grass_park.png"
	SpriteTileCobblePlaza    = "sprites/map/tile_cobble_plaza.png"
	SpriteTileBoardwalk      = "sprites/map/tile_boardwalk.png"
	SpriteTileCanalWater     = "sprites/map/tile_canal_water.png"
	SpriteTilePlazaMosaic    = "sprites/map/tile_plaza_mosaic.png"
	SpriteTileWetAsphalt     = "sprites/map/tile_wet_asphalt.png"
	SpriteTileParkPath       = "sprites/map/tile_park_path.png"
	SpriteTileBikeLane       = "sprites/map/tile_bike_lane.png"
	SpriteTileCanalEdge      = "sprites/map/tile_canal_edge.png"

	// UI & Pickup Icons
	SpriteIconUpgCooldown     = "sprites/ui/icon_upg_cooldown.png"
	SpriteIconUpgMultiHook    = "sprites/ui/icon_upg_multihook.png"
	SpriteIconUpgMaxStops     = "sprites/ui/icon_upg_maxstops.png"
	SpriteIconUpgRange        = "sprites/ui/icon_upg_range.png"
	SpriteIconUpgChronos      = "sprites/ui/icon_upg_chronos.png"
	SpriteIconUpgVitality     = "sprites/ui/icon_upg_vitality.png"
	SpriteIconRelicHeadphones = "sprites/ui/icon_relic_headphones.png"
	SpriteIconRelicPass       = "sprites/ui/icon_relic_pass.png"
	SpriteIconRelicShield     = "sprites/ui/icon_relic_shield.png"
	SpriteIconRelicClover     = "sprites/ui/icon_relic_clover.png"
	SpriteIconRelicEspresso   = "sprites/ui/icon_relic_espresso.png"
	SpriteIconRelicVIP        = "sprites/ui/icon_relic_vip.png"
	SpriteIconShopEspresso    = "sprites/ui/icon_shop_espresso.png"
	SpriteIconShopBribe       = "sprites/ui/icon_shop_bribe.png"
	SpriteIconShopRelic       = "sprites/ui/icon_shop_relic.png"
	SpriteIconShopCrate       = "sprites/ui/icon_shop_crate.png"
	SpriteIconSlotCherry      = "sprites/ui/icon_slot_cherry.png"
	SpriteIconSlotCash        = "sprites/ui/icon_slot_cash.png"
	SpriteIconSlotXP          = "sprites/ui/icon_slot_xp.png"
	SpriteIconSlotHeart       = "sprites/ui/icon_slot_heart.png"
	SpriteIconSlot777         = "sprites/ui/icon_slot_777.png"
	SpriteIconSlotSiren       = "sprites/ui/icon_slot_siren.png"
	SpriteIconMouseLMB        = "sprites/ui/icon_mouse_lmb.png"
	SpriteIconMouseRMB        = "sprites/ui/icon_mouse_rmb.png"
	SpriteIconMouseWheel      = "sprites/ui/icon_mouse_wheel.png"
	SpriteIconHeartFull       = "sprites/ui/icon_heart_full.png"
	SpriteIconHeartEmpty      = "sprites/ui/icon_heart_empty.png"
	SpriteIconStarWanted      = "sprites/ui/icon_star_wanted.png"
	SpriteIconCoin            = "sprites/ui/icon_coin.png"
	SpriteIconBolt            = "sprites/ui/icon_bolt.png"
	SpriteIconPickupHeal      = "sprites/ui/icon_pickup_heal.png"
)

type spriteSpec struct {
	path  string
	w, h  int
	fill  color.RGBA
	acc   color.RGBA
	shape string
}

var expectedSpecs = []spriteSpec{
	// Alfredo
	{SpriteAlfredoIdle0, 32, 32, color.RGBA{46, 90, 136, 255}, color.RGBA{255, 224, 200, 255}, "char"},
	{SpriteAlfredoIdle1, 32, 32, color.RGBA{46, 90, 136, 255}, color.RGBA{255, 224, 200, 255}, "char"},
	{SpriteAlfredoRun0, 32, 32, color.RGBA{46, 90, 136, 255}, color.RGBA{230, 57, 70, 255}, "char"},
	{SpriteAlfredoRun1, 32, 32, color.RGBA{46, 90, 136, 255}, color.RGBA{230, 57, 70, 255}, "char"},
	{SpriteAlfredoRun2, 32, 32, color.RGBA{46, 90, 136, 255}, color.RGBA{230, 57, 70, 255}, "char"},
	{SpriteAlfredoRun3, 32, 32, color.RGBA{46, 90, 136, 255}, color.RGBA{230, 57, 70, 255}, "char"},
	{SpriteAlfredoDash, 32, 32, color.RGBA{0, 245, 212, 255}, color.RGBA{255, 224, 200, 255}, "char"},
	{SpriteAlfredoThrow, 32, 32, color.RGBA{75, 127, 179, 255}, color.RGBA{255, 224, 200, 255}, "char"},

	// Alfredo skin: Neon Drip
	{SpriteAlfredoDripIdle0, 32, 32, color.RGBA{255, 42, 133, 255}, color.RGBA{0, 240, 255, 255}, "char"},
	{SpriteAlfredoDripIdle1, 32, 32, color.RGBA{255, 42, 133, 255}, color.RGBA{0, 240, 255, 255}, "char"},
	{SpriteAlfredoDripRun0, 32, 32, color.RGBA{255, 42, 133, 255}, color.RGBA{0, 240, 255, 255}, "char"},
	{SpriteAlfredoDripRun1, 32, 32, color.RGBA{255, 42, 133, 255}, color.RGBA{0, 240, 255, 255}, "char"},
	{SpriteAlfredoDripRun2, 32, 32, color.RGBA{255, 42, 133, 255}, color.RGBA{0, 240, 255, 255}, "char"},
	{SpriteAlfredoDripRun3, 32, 32, color.RGBA{255, 42, 133, 255}, color.RGBA{0, 240, 255, 255}, "char"},
	{SpriteAlfredoDripDash, 32, 32, color.RGBA{255, 42, 133, 255}, color.RGBA{0, 240, 255, 255}, "char"},
	{SpriteAlfredoDripThrow, 32, 32, color.RGBA{255, 42, 133, 255}, color.RGBA{0, 240, 255, 255}, "char"},

	// Alfredo skin: Alfredo Banana
	{SpriteAlfredoBananaIdle0, 32, 32, color.RGBA{255, 225, 53, 255}, color.RGBA{255, 224, 200, 255}, "char"},
	{SpriteAlfredoBananaIdle1, 32, 32, color.RGBA{255, 225, 53, 255}, color.RGBA{255, 224, 200, 255}, "char"},
	{SpriteAlfredoBananaRun0, 32, 32, color.RGBA{255, 225, 53, 255}, color.RGBA{255, 224, 200, 255}, "char"},
	{SpriteAlfredoBananaRun1, 32, 32, color.RGBA{255, 225, 53, 255}, color.RGBA{255, 224, 200, 255}, "char"},
	{SpriteAlfredoBananaRun2, 32, 32, color.RGBA{255, 225, 53, 255}, color.RGBA{255, 224, 200, 255}, "char"},
	{SpriteAlfredoBananaRun3, 32, 32, color.RGBA{255, 225, 53, 255}, color.RGBA{255, 224, 200, 255}, "char"},
	{SpriteAlfredoBananaDash, 32, 32, color.RGBA{255, 225, 53, 255}, color.RGBA{0, 240, 255, 255}, "char"},
	{SpriteAlfredoBananaThrow, 32, 32, color.RGBA{255, 225, 53, 255}, color.RGBA{255, 224, 200, 255}, "char"},

	// Alfredo skin: Aletta di Pollo
	{SpriteAlfredoChickenIdle0, 32, 32, color.RGBA{217, 138, 48, 255}, color.RGBA{255, 243, 232, 255}, "char"},
	{SpriteAlfredoChickenIdle1, 32, 32, color.RGBA{217, 138, 48, 255}, color.RGBA{255, 243, 232, 255}, "char"},
	{SpriteAlfredoChickenRun0, 32, 32, color.RGBA{217, 138, 48, 255}, color.RGBA{255, 243, 232, 255}, "char"},
	{SpriteAlfredoChickenRun1, 32, 32, color.RGBA{217, 138, 48, 255}, color.RGBA{255, 243, 232, 255}, "char"},
	{SpriteAlfredoChickenRun2, 32, 32, color.RGBA{217, 138, 48, 255}, color.RGBA{255, 243, 232, 255}, "char"},
	{SpriteAlfredoChickenRun3, 32, 32, color.RGBA{217, 138, 48, 255}, color.RGBA{255, 243, 232, 255}, "char"},
	{SpriteAlfredoChickenDash, 32, 32, color.RGBA{217, 138, 48, 255}, color.RGBA{0, 240, 255, 255}, "char"},
	{SpriteAlfredoChickenThrow, 32, 32, color.RGBA{217, 138, 48, 255}, color.RGBA{255, 243, 232, 255}, "char"},

	// Alfredo skin: Cyber-Samurai Zero
	{SpriteAlfredoSamuraiIdle0, 32, 32, color.RGBA{184, 29, 58, 255}, color.RGBA{255, 209, 102, 255}, "char"},
	{SpriteAlfredoSamuraiIdle1, 32, 32, color.RGBA{184, 29, 58, 255}, color.RGBA{255, 209, 102, 255}, "char"},
	{SpriteAlfredoSamuraiRun0, 32, 32, color.RGBA{184, 29, 58, 255}, color.RGBA{255, 209, 102, 255}, "char"},
	{SpriteAlfredoSamuraiRun1, 32, 32, color.RGBA{184, 29, 58, 255}, color.RGBA{255, 209, 102, 255}, "char"},
	{SpriteAlfredoSamuraiRun2, 32, 32, color.RGBA{184, 29, 58, 255}, color.RGBA{255, 209, 102, 255}, "char"},
	{SpriteAlfredoSamuraiRun3, 32, 32, color.RGBA{184, 29, 58, 255}, color.RGBA{255, 209, 102, 255}, "char"},
	{SpriteAlfredoSamuraiDash, 32, 32, color.RGBA{184, 29, 58, 255}, color.RGBA{0, 240, 255, 255}, "char"},
	{SpriteAlfredoSamuraiThrow, 32, 32, color.RGBA{184, 29, 58, 255}, color.RGBA{255, 209, 102, 255}, "char"},

	// Girls
	{SpriteGirlKawaii0, 32, 32, color.RGBA{255, 107, 157, 255}, color.RGBA{255, 240, 245, 255}, "char"},
	{SpriteGirlKawaiiSheet, 96, 32, color.RGBA{255, 107, 157, 255}, color.RGBA{255, 240, 245, 255}, "sheet"},
	{SpriteGirlEmo0, 32, 32, color.RGBA{109, 59, 142, 255}, color.RGBA{224, 82, 136, 255}, "char"},
	{SpriteGirlEmoSheet, 96, 32, color.RGBA{109, 59, 142, 255}, color.RGBA{224, 82, 136, 255}, "sheet"},
	{SpriteGirlStudent0, 32, 32, color.RGBA{59, 105, 152, 255}, color.RGBA{217, 138, 78, 255}, "char"},
	{SpriteGirlStudentSheet, 96, 32, color.RGBA{59, 105, 152, 255}, color.RGBA{217, 138, 78, 255}, "sheet"},
	{SpriteGirlGyaru0, 32, 32, color.RGBA{255, 184, 77, 255}, color.RGBA{0, 240, 255, 255}, "char"},
	{SpriteGirlGyaruSheet, 96, 32, color.RGBA{255, 184, 77, 255}, color.RGBA{0, 240, 255, 255}, "sheet"},
	{SpriteVIPAura, 32, 32, color.RGBA{255, 230, 0, 140}, color.RGBA{255, 255, 180, 220}, "ring"},

	// Stops & Shop
	{SpriteStopKawaii, 48, 48, color.RGBA{255, 107, 157, 255}, color.RGBA{255, 240, 245, 255}, "stop"},
	{SpriteStopEmo, 48, 48, color.RGBA{109, 59, 142, 255}, color.RGBA{224, 82, 136, 255}, "stop"},
	{SpriteStopStudent, 48, 48, color.RGBA{59, 105, 152, 255}, color.RGBA{0, 245, 212, 255}, "stop"},
	{SpriteStopGyaru, 48, 48, color.RGBA{255, 184, 77, 255}, color.RGBA{0, 240, 255, 255}, "stop"},
	{SpriteShopVending, 32, 32, color.RGBA{220, 48, 65, 255}, color.RGBA{0, 245, 212, 255}, "stop"},

	// Bus & Police
	{SpriteBusNeon, 96, 48, color.RGBA{255, 42, 133, 255}, color.RGBA{0, 245, 212, 255}, "bus"},
	{SpriteBusGold, 96, 48, color.RGBA{255, 209, 102, 255}, color.RGBA{230, 57, 70, 255}, "bus"},
	{SpritePoliceOfficer, 32, 32, color.RGBA{36, 67, 140, 255}, color.RGBA{255, 60, 60, 255}, "char"},
	{SpritePoliceCar, 48, 32, color.RGBA{30, 50, 110, 255}, color.RGBA{255, 40, 60, 255}, "bus"},

	// Bosses
	{SpriteBossBestFriend, 36, 36, color.RGBA{255, 42, 133, 255}, color.RGBA{255, 230, 0, 255}, "char"},
	{SpriteBossLover, 36, 36, color.RGBA{200, 30, 80, 255}, color.RGBA{255, 110, 175, 255}, "char"},
	{SpriteBossProfessor, 36, 36, color.RGBA{75, 105, 165, 255}, color.RGBA{255, 230, 0, 255}, "char"},
	{SpriteBossExecutive, 36, 36, color.RGBA{38, 56, 98, 255}, color.RGBA{255, 210, 90, 255}, "char"},
	{SpriteBossRector, 36, 36, color.RGBA{112, 48, 160, 255}, color.RGBA{255, 224, 100, 255}, "char"},
	{SpriteBossMom, 36, 36, color.RGBA{180, 60, 90, 255}, color.RGBA{255, 210, 120, 255}, "char"},

	// Bullets
	{SpriteBulletRazor, 16, 16, color.RGBA{224, 82, 136, 255}, color.RGBA{255, 255, 255, 255}, "bullet"},
	{SpriteBulletProtractor, 16, 16, color.RGBA{255, 230, 0, 255}, color.RGBA{255, 255, 255, 255}, "bullet"},
	{SpriteBulletTear, 16, 16, color.RGBA{0, 240, 255, 255}, color.RGBA{255, 255, 255, 255}, "bullet"},
	{SpriteBulletFlash, 16, 16, color.RGBA{255, 255, 100, 255}, color.RGBA{0, 240, 255, 255}, "bullet"},
	{SpriteBulletPolice, 16, 16, color.RGBA{255, 50, 70, 255}, color.RGBA{60, 140, 255, 255}, "bullet"},
	{SpriteBulletCiabatta, 16, 16, color.RGBA{235, 120, 60, 255}, color.RGBA{255, 240, 200, 255}, "bullet"},
	{SpriteBulletLipstick, 16, 16, color.RGBA{255, 42, 133, 255}, color.RGBA{255, 210, 230, 255}, "bullet"},
	{SpriteBulletContract, 16, 16, color.RGBA{240, 235, 225, 255}, color.RGBA{230, 55, 75, 255}, "bullet"},
	{SpriteBulletSeal, 16, 16, color.RGBA{255, 210, 60, 255}, color.RGBA{190, 40, 80, 255}, "bullet"},

	// Tiles
	{SpriteTileAsphalt, 32, 32, color.RGBA{88, 85, 104, 255}, color.RGBA{63, 59, 77, 255}, "tile"},
	{SpriteTileSidewalk, 32, 32, color.RGBA{200, 191, 168, 255}, color.RGBA{172, 163, 140, 255}, "tile"},
	{SpriteTileZTL, 32, 32, color.RGBA{120, 42, 68, 255}, color.RGBA{255, 42, 133, 255}, "tile"},
	{SpriteTileCrosswalk, 32, 32, color.RGBA{88, 85, 104, 255}, color.RGBA{235, 235, 240, 255}, "crosswalk"},

	// Layered Map Assets
	{SpriteBuildingKonbini, 96, 64, color.RGBA{235, 224, 208, 255}, color.RGBA{255, 107, 138, 255}, "building"},
	{SpriteBuildingCoffee, 96, 64, color.RGBA{196, 146, 102, 255}, color.RGBA{255, 216, 138, 255}, "building"},
	{SpriteBuildingArcade, 96, 64, color.RGBA{62, 54, 108, 255}, color.RGBA{0, 245, 212, 255}, "building"},
	{SpriteBuildingRamen, 96, 64, color.RGBA{152, 68, 52, 255}, color.RGBA{255, 90, 80, 255}, "building"},
	{SpriteBuildingBookstore, 96, 64, color.RGBA{68, 122, 134, 255}, color.RGBA{255, 206, 98, 255}, "building"},
	{SpriteBuildingMetro, 96, 64, color.RGBA{58, 72, 102, 255}, color.RGBA{0, 225, 255, 255}, "building"},
	{SpriteBuildingNewsstand, 48, 40, color.RGBA{228, 214, 196, 255}, color.RGBA{255, 92, 118, 255}, "building"},
	{SpriteBuildingPizzeria, 96, 64, color.RGBA{214, 92, 68, 255}, color.RGBA{255, 226, 122, 255}, "building"},
	{SpriteBuildingVinylBar, 96, 64, color.RGBA{72, 46, 114, 255}, color.RGBA{255, 152, 224, 255}, "building"},
	{SpriteBuildingGelateria, 96, 64, color.RGBA{244, 198, 218, 255}, color.RGBA{112, 240, 214, 255}, "building"},
	{SpritePropFountain, 64, 48, color.RGBA{182, 188, 204, 255}, color.RGBA{60, 235, 255, 255}, "stop"},
	{SpritePropArchGate, 96, 40, color.RGBA{210, 68, 92, 255}, color.RGBA{255, 214, 112, 255}, "building"},
	{SpritePropTreeLush, 48, 48, color.RGBA{88, 178, 92, 255}, color.RGBA{148, 218, 118, 255}, "tree"},
	{SpritePropStreetlamp, 24, 48, color.RGBA{56, 52, 76, 255}, color.RGBA{255, 216, 138, 255}, "lamp"},
	{SpritePropBenchPlants, 48, 32, color.RGBA{178, 118, 72, 255}, color.RGBA{92, 176, 98, 255}, "bench"},
	{SpritePropCatsClutter, 32, 32, color.RGBA{192, 148, 98, 255}, color.RGBA{255, 142, 172, 255}, "cats"},
	{SpritePropScooter, 32, 24, color.RGBA{105, 222, 188, 255}, color.RGBA{255, 244, 224, 255}, "car"},
	{SpritePropBikeRack, 32, 20, color.RGBA{130, 145, 165, 255}, color.RGBA{255, 110, 145, 255}, "bench"},
	{SpritePropHydrantBin, 24, 24, color.RGBA{68, 128, 178, 255}, color.RGBA{230, 62, 74, 255}, "stop"},
	{SpritePropZTLSign, 24, 32, color.RGBA{64, 68, 90, 255}, color.RGBA{255, 50, 85, 255}, "lamp"},
	{SpritePropCanalBoat, 48, 24, color.RGBA{168, 106, 62, 255}, color.RGBA{255, 216, 138, 255}, "bus"},
	{SpritePropSlotMachine, 32, 32, color.RGBA{112, 38, 148, 255}, color.RGBA{255, 225, 48, 255}, "stop"},
	{SpritePropManhole, 24, 24, color.RGBA{70, 66, 84, 255}, color.RGBA{96, 92, 112, 255}, "tile"},
	{SpritePropAsphaltCrack, 32, 16, color.RGBA{70, 66, 84, 255}, color.RGBA{96, 92, 112, 255}, "tile"},
	{SpritePropFlowerBox, 32, 20, color.RGBA{164, 108, 66, 255}, color.RGBA{255, 120, 170, 255}, "bench"},
	{SpritePropNeonBillboard, 48, 32, color.RGBA{34, 28, 64, 255}, color.RGBA{0, 245, 212, 255}, "building"},
	{SpritePropPhoneBooth, 24, 40, color.RGBA{222, 54, 96, 255}, color.RGBA{0, 245, 212, 255}, "stop"},
	{SpritePropFoodCart, 48, 36, color.RGBA{220, 128, 66, 255}, color.RGBA{255, 218, 118, 255}, "building"},
	{SpritePropKioskUmbrella, 40, 36, color.RGBA{242, 102, 142, 255}, color.RGBA{255, 244, 224, 255}, "stop"},
	{SpritePropTrafficLight, 24, 48, color.RGBA{56, 58, 76, 255}, color.RGBA{70, 245, 140, 255}, "lamp"},
	{SpritePropBollardGlow, 16, 24, color.RGBA{64, 68, 92, 255}, color.RGBA{0, 245, 212, 255}, "lamp"},
	{SpritePropPlanterNeon, 48, 24, color.RGBA{92, 176, 104, 255}, color.RGBA{255, 198, 112, 255}, "bench"},
	{SpritePropVendingDrink, 28, 36, color.RGBA{64, 176, 232, 255}, color.RGBA{215, 248, 255, 255}, "stop"},
	{SpritePropBusShelterDeco, 48, 36, color.RGBA{72, 92, 134, 255}, color.RGBA{255, 222, 148, 255}, "building"},
	{SpriteCarRed, 48, 32, color.RGBA{222, 68, 86, 255}, color.RGBA{185, 230, 255, 255}, "car"},
	{SpriteCarYellow, 48, 32, color.RGBA{242, 198, 62, 255}, color.RGBA{185, 230, 255, 255}, "car"},
	{SpriteCarMint, 48, 32, color.RGBA{84, 212, 178, 255}, color.RGBA{215, 248, 255, 255}, "car"},
	{SpriteCarVanCyber, 64, 36, color.RGBA{68, 78, 128, 255}, color.RGBA{0, 245, 212, 255}, "car"},
	{SpriteTileGrassPark, 32, 32, color.RGBA{98, 172, 92, 255}, color.RGBA{128, 198, 108, 255}, "grass"},
	{SpriteTileCobblePlaza, 32, 32, color.RGBA{182, 168, 152, 255}, color.RGBA{142, 128, 116, 255}, "cobble"},
	{SpriteTileBoardwalk, 32, 32, color.RGBA{168, 118, 76, 255}, color.RGBA{124, 82, 48, 255}, "boardwalk"},
	{SpriteTileCanalWater, 32, 32, color.RGBA{42, 128, 168, 255}, color.RGBA{112, 210, 235, 255}, "water"},
	{SpriteTilePlazaMosaic, 32, 32, color.RGBA{196, 182, 166, 255}, color.RGBA{146, 132, 152, 255}, "cobble"},
	{SpriteTileWetAsphalt, 32, 32, color.RGBA{72, 72, 96, 255}, color.RGBA{108, 126, 164, 255}, "tile"},
	{SpriteTileParkPath, 32, 32, color.RGBA{192, 176, 148, 255}, color.RGBA{154, 138, 112, 255}, "cobble"},
	{SpriteTileBikeLane, 32, 32, color.RGBA{74, 124, 118, 255}, color.RGBA{235, 242, 245, 255}, "tile"},
	{SpriteTileCanalEdge, 32, 32, color.RGBA{152, 142, 134, 255}, color.RGBA{72, 148, 182, 255}, "boardwalk"},

	// UI & Pickup Icons
	{SpriteIconUpgCooldown, 20, 20, color.RGBA{0, 245, 212, 255}, color.RGBA{255, 255, 255, 255}, "bullet"},
	{SpriteIconUpgMultiHook, 20, 20, color.RGBA{255, 107, 157, 255}, color.RGBA{255, 230, 0, 255}, "bullet"},
	{SpriteIconUpgMaxStops, 20, 20, color.RGBA{255, 184, 77, 255}, color.RGBA{0, 245, 212, 255}, "bullet"},
	{SpriteIconUpgRange, 20, 20, color.RGBA{92, 186, 255, 255}, color.RGBA{255, 255, 255, 255}, "bullet"},
	{SpriteIconUpgChronos, 20, 20, color.RGBA{176, 98, 255, 255}, color.RGBA{255, 230, 120, 255}, "bullet"},
	{SpriteIconUpgVitality, 20, 20, color.RGBA{255, 68, 108, 255}, color.RGBA{255, 224, 236, 255}, "bullet"},
	{SpriteIconRelicHeadphones, 16, 16, color.RGBA{255, 42, 133, 255}, color.RGBA{0, 245, 212, 255}, "bullet"},
	{SpriteIconRelicPass, 16, 16, color.RGBA{255, 210, 68, 255}, color.RGBA{64, 176, 255, 255}, "bullet"},
	{SpriteIconRelicShield, 16, 16, color.RGBA{64, 210, 255, 255}, color.RGBA{255, 255, 255, 255}, "bullet"},
	{SpriteIconRelicClover, 16, 16, color.RGBA{88, 222, 118, 255}, color.RGBA{210, 255, 180, 255}, "bullet"},
	{SpriteIconRelicEspresso, 16, 16, color.RGBA{168, 96, 58, 255}, color.RGBA{255, 222, 150, 255}, "bullet"},
	{SpriteIconRelicVIP, 16, 16, color.RGBA{255, 215, 48, 255}, color.RGBA{255, 92, 168, 255}, "bullet"},
	{SpriteIconShopEspresso, 16, 16, color.RGBA{182, 102, 64, 255}, color.RGBA{255, 232, 184, 255}, "bullet"},
	{SpriteIconShopBribe, 16, 16, color.RGBA{86, 214, 122, 255}, color.RGBA{255, 230, 96, 255}, "bullet"},
	{SpriteIconShopRelic, 16, 16, color.RGBA{182, 92, 255, 255}, color.RGBA{255, 224, 112, 255}, "bullet"},
	{SpriteIconShopCrate, 16, 16, color.RGBA{214, 146, 78, 255}, color.RGBA{0, 245, 212, 255}, "bullet"},
	{SpriteIconSlotCherry, 20, 20, color.RGBA{235, 52, 86, 255}, color.RGBA{108, 224, 112, 255}, "bullet"},
	{SpriteIconSlotCash, 20, 20, color.RGBA{82, 212, 118, 255}, color.RGBA{255, 230, 96, 255}, "bullet"},
	{SpriteIconSlotXP, 20, 20, color.RGBA{72, 198, 255, 255}, color.RGBA{255, 255, 255, 255}, "bullet"},
	{SpriteIconSlotHeart, 20, 20, color.RGBA{255, 68, 126, 255}, color.RGBA{255, 218, 234, 255}, "bullet"},
	{SpriteIconSlot777, 20, 20, color.RGBA{255, 210, 48, 255}, color.RGBA{255, 64, 132, 255}, "bullet"},
	{SpriteIconSlotSiren, 20, 20, color.RGBA{255, 48, 72, 255}, color.RGBA{72, 152, 255, 255}, "bullet"},
	{SpriteIconMouseLMB, 16, 16, color.RGBA{210, 214, 228, 255}, color.RGBA{0, 245, 212, 255}, "stop"},
	{SpriteIconMouseRMB, 16, 16, color.RGBA{210, 214, 228, 255}, color.RGBA{255, 42, 133, 255}, "stop"},
	{SpriteIconMouseWheel, 16, 16, color.RGBA{210, 214, 228, 255}, color.RGBA{255, 216, 72, 255}, "stop"},
	{SpriteIconHeartFull, 12, 12, color.RGBA{255, 58, 102, 255}, color.RGBA{255, 214, 228, 255}, "bullet"},
	{SpriteIconHeartEmpty, 12, 12, color.RGBA{74, 62, 92, 255}, color.RGBA{128, 114, 148, 255}, "bullet"},
	{SpriteIconStarWanted, 12, 12, color.RGBA{255, 216, 48, 255}, color.RGBA{255, 86, 72, 255}, "bullet"},
	{SpriteIconCoin, 12, 12, color.RGBA{255, 212, 56, 255}, color.RGBA{255, 246, 178, 255}, "bullet"},
	{SpriteIconBolt, 12, 12, color.RGBA{0, 245, 212, 255}, color.RGBA{255, 240, 120, 255}, "bullet"},
	{SpriteIconPickupHeal, 16, 16, color.RGBA{176, 98, 62, 255}, color.RGBA{255, 88, 128, 255}, "bullet"},
}

// SpriteBank caches all embedded sprites and guarantees crisp fallback images for any missing asset.
type SpriteBank struct {
	images map[string]*ebiten.Image
	frames map[string][]*ebiten.Image
}

// LoadSpriteBank reads embedded PNGs and builds fallbacks for any missing expected asset.
func LoadSpriteBank() *SpriteBank {
	bank := &SpriteBank{
		images: make(map[string]*ebiten.Image, len(expectedSpecs)+8),
		frames: make(map[string][]*ebiten.Image, 8),
	}

	// Preload everything found under embedded sprites/*/*.png (skipping benchmark and preview_* sheets)
	if entries, err := spriteFS.ReadDir("sprites"); err == nil {
		for _, dir := range entries {
			if !dir.IsDir() || dir.Name() == "benchmark" {
				continue
			}
			subPath := path.Join("sprites", dir.Name())
			files, err := spriteFS.ReadDir(subPath)
			if err != nil {
				continue
			}
			for _, f := range files {
				name := strings.ToLower(f.Name())
				if f.IsDir() || !strings.HasSuffix(name, ".png") || strings.HasPrefix(name, "preview_") {
					continue
				}
				fullPath := path.Join(subPath, f.Name())
				if img := loadPNGFromFS(fullPath); img != nil {
					bank.images[fullPath] = img
				}
			}
		}
	}

	// Ensure every expected key exists, generating a crisp fallback if missing.
	for _, spec := range expectedSpecs {
		if _, ok := bank.images[spec.path]; !ok {
			bank.images[spec.path] = makeFallbackImage(spec)
		}
	}

	return bank
}

// PlayerSkinSpriteKey returns the embedded sprite key for a player skin (0..4) and pose name.
func PlayerSkinSpriteKey(skin int, pose string) string {
	if pose == "" {
		pose = "idle_0"
	}
	var prefix string
	switch skin {
	case 1:
		prefix = "alfredo_drip_"
	case 2:
		prefix = "alfredo_banana_"
	case 3:
		prefix = "alfredo_chicken_"
	case 4:
		prefix = "alfredo_samurai_"
	default:
		prefix = "alfredo_"
	}
	return "sprites/characters/" + prefix + pose + ".png"
}

func loadPNGFromFS(p string) *ebiten.Image {
	data, err := spriteFS.ReadFile(p)
	if err != nil {
		return nil
	}
	decoded, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil
	}
	return ebiten.NewImageFromImage(decoded)
}

// Get returns a loaded or fallback sprite for key. If an unknown key is requested, it creates a 32x32 fallback.
func (sb *SpriteBank) Get(key string) *ebiten.Image {
	if img, ok := sb.images[key]; ok && img != nil {
		return img
	}
	if !strings.HasPrefix(key, "sprites/") {
		if img, ok := sb.images["sprites/"+key]; ok && img != nil {
			return img
		}
	}
	fb := makeFallbackImage(spriteSpec{
		path:  key,
		w:     32,
		h:     32,
		fill:  color.RGBA{255, 0, 255, 255},
		acc:   color.RGBA{0, 245, 212, 255},
		shape: "char",
	})
	sb.images[key] = fb
	return fb
}

// GetOrFallback returns the sprite for key or builds a fallback of the exact requested dimensions if missing.
func (sb *SpriteBank) GetOrFallback(key string, w, h int) *ebiten.Image {
	if img, ok := sb.images[key]; ok && img != nil {
		return img
	}
	if !strings.HasPrefix(key, "sprites/") {
		if img, ok := sb.images["sprites/"+key]; ok && img != nil {
			return img
		}
	}
	if w <= 0 {
		w = 32
	}
	if h <= 0 {
		h = 32
	}
	fb := makeFallbackImage(spriteSpec{
		path:  key,
		w:     w,
		h:     h,
		fill:  color.RGBA{255, 42, 133, 255},
		acc:   color.RGBA{0, 245, 212, 255},
		shape: "char",
	})
	sb.images[key] = fb
	return fb
}

// GetSheetFrame extracts and caches a frame from a horizontal sprite sheet.
func (sb *SpriteBank) GetSheetFrame(sheetKey string, frameIdx, frameW, frameH int) *ebiten.Image {
	if frames, ok := sb.frames[sheetKey]; ok && len(frames) > 0 {
		if frameIdx < 0 {
			frameIdx = 0
		}
		return frames[frameIdx%len(frames)]
	}

	sheet := sb.GetOrFallback(sheetKey, frameW*4, frameH)
	b := sheet.Bounds()
	w := b.Dx()
	h := b.Dy()
	if frameW <= 0 {
		frameW = h
	}
	if frameH <= 0 {
		frameH = h
	}
	count := w / frameW
	if count < 1 {
		sb.frames[sheetKey] = []*ebiten.Image{sheet}
		return sheet
	}

	sliced := make([]*ebiten.Image, 0, count)
	for i := 0; i < count; i++ {
		rect := image.Rect(b.Min.X+i*frameW, b.Min.Y, b.Min.X+(i+1)*frameW, b.Min.Y+frameH)
		sub, ok := sheet.SubImage(rect).(*ebiten.Image)
		if !ok {
			sub = sheet
		}
		sliced = append(sliced, sub)
	}
	sb.frames[sheetKey] = sliced
	if frameIdx < 0 {
		frameIdx = 0
	}
	return sliced[frameIdx%len(sliced)]
}

func makeFallbackImage(spec spriteSpec) *ebiten.Image {
	rgba := image.NewRGBA(image.Rect(0, 0, spec.w, spec.h))
	outline := color.RGBA{28, 17, 36, 255} // #1C1124 Deep Aubergine

	switch spec.shape {
	case "tile":
		for y := 0; y < spec.h; y++ {
			for x := 0; x < spec.w; x++ {
				c := spec.fill
				if (x == 0 || y == 0) && (x+y)%2 == 0 {
					c = spec.acc
				} else if (x == 7 || x == 23) && (y == 11 || y == 27) {
					c = spec.acc
				}
				rgba.SetRGBA(x, y, c)
			}
		}
	case "grass":
		for y := 0; y < spec.h; y++ {
			for x := 0; x < spec.w; x++ {
				c := spec.fill
				if ((x*5+y*3)%19 == 0) || ((x+y)%13 == 0 && y%2 == 0) {
					c = spec.acc
				}
				rgba.SetRGBA(x, y, c)
			}
		}
	case "cobble":
		for y := 0; y < spec.h; y++ {
			for x := 0; x < spec.w; x++ {
				row := y / 8
				ox := (row % 2) * 4
				if y%8 == 0 || (x+ox)%8 == 0 {
					rgba.SetRGBA(x, y, spec.acc)
				} else {
					rgba.SetRGBA(x, y, spec.fill)
				}
			}
		}
	case "boardwalk":
		for y := 0; y < spec.h; y++ {
			for x := 0; x < spec.w; x++ {
				if y%6 == 0 || (x == 15 && (y/6)%2 == 0) || (x == 31 && (y/6)%2 == 1) {
					rgba.SetRGBA(x, y, spec.acc)
				} else {
					rgba.SetRGBA(x, y, spec.fill)
				}
			}
		}
	case "water":
		for y := 0; y < spec.h; y++ {
			for x := 0; x < spec.w; x++ {
				c := spec.fill
				if (y%7 == 2 && (x+y*2)%11 < 4) || (y%11 == 6 && x%9 < 3) {
					c = spec.acc
				}
				rgba.SetRGBA(x, y, c)
			}
		}
	case "crosswalk":
		for y := 0; y < spec.h; y++ {
			for x := 0; x < spec.w; x++ {
				if (y/6)%2 == 0 && x >= 4 && x < spec.w-4 {
					rgba.SetRGBA(x, y, spec.acc)
				} else {
					rgba.SetRGBA(x, y, spec.fill)
				}
			}
		}
	case "building":
		drawBoxFallback(rgba, 2, 2, spec.w-4, spec.h-4, spec.fill, spec.acc, outline)
		// Warm glowing storefront window
		winCol := color.RGBA{255, 224, 150, 255}
		for y := spec.h - 22; y < spec.h-8; y++ {
			for x := 12; x < spec.w-12; x++ {
				if x == 12 || x == spec.w-13 || y == spec.h-22 || y == spec.h-9 {
					rgba.SetRGBA(x, y, outline)
				} else if (x-12)%24 < 18 {
					rgba.SetRGBA(x, y, winCol)
				}
			}
		}
	case "tree":
		cx, cy := spec.w/2, spec.h/2-4
		trunkCol := color.RGBA{128, 82, 52, 255}
		for y := spec.h - 16; y < spec.h-2; y++ {
			for x := cx - 4; x <= cx+4; x++ {
				if x == cx-4 || x == cx+4 || y == spec.h-3 {
					rgba.SetRGBA(x, y, outline)
				} else {
					rgba.SetRGBA(x, y, trunkCol)
				}
			}
		}
		r := spec.w/2 - 5
		for y := 2; y < spec.h-10; y++ {
			for x := 2; x < spec.w-2; x++ {
				dx, dy := x-cx, y-cy
				d2 := dx*dx + dy*dy
				if d2 <= r*r {
					c := spec.fill
					if dx+dy < -4 {
						c = spec.acc
					}
					rgba.SetRGBA(x, y, c)
				} else if d2 <= (r+1)*(r+1) {
					rgba.SetRGBA(x, y, outline)
				}
			}
		}
	case "lamp":
		cx := spec.w / 2
		for y := 10; y < spec.h-2; y++ {
			for x := cx - 2; x <= cx+2; x++ {
				if x == cx-2 || x == cx+2 || y == spec.h-3 {
					rgba.SetRGBA(x, y, outline)
				} else {
					rgba.SetRGBA(x, y, spec.fill)
				}
			}
		}
		for y := 3; y < 12; y++ {
			for x := cx - 5; x <= cx+5; x++ {
				if x == cx-5 || x == cx+5 || y == 3 || y == 11 {
					rgba.SetRGBA(x, y, outline)
				} else {
					rgba.SetRGBA(x, y, spec.acc)
				}
			}
		}
	case "bench":
		drawBoxFallback(rgba, 6, 10, spec.w-12, spec.h-14, spec.fill, color.RGBA{214, 152, 98, 255}, outline)
		// Potted plant on right
		drawBoxFallback(rgba, spec.w-14, 4, 10, 14, spec.acc, color.RGBA{148, 218, 118, 255}, outline)
	case "cats":
		drawBoxFallback(rgba, 4, 10, 14, 16, spec.fill, color.RGBA{222, 178, 128, 255}, outline)
		drawBoxFallback(rgba, 18, 14, 10, 12, spec.acc, color.RGBA{255, 220, 232, 255}, outline)
	case "car":
		drawBoxFallback(rgba, 2, 6, spec.w-4, spec.h-10, spec.fill, spec.acc, outline)
	case "bullet":
		cx, cy := spec.w/2, spec.h/2
		r := spec.w/2 - 2
		for y := 0; y < spec.h; y++ {
			for x := 0; x < spec.w; x++ {
				dx, dy := x-cx, y-cy
				d2 := dx*dx + dy*dy
				if d2 <= (r-2)*(r-2) {
					rgba.SetRGBA(x, y, spec.acc)
				} else if d2 <= r*r {
					rgba.SetRGBA(x, y, spec.fill)
				} else if d2 <= (r+1)*(r+1) {
					rgba.SetRGBA(x, y, outline)
				}
			}
		}
	case "ring":
		cx, cy := spec.w/2, spec.h/2
		rOut := spec.w/2 - 2
		rIn := rOut - 3
		for y := 0; y < spec.h; y++ {
			for x := 0; x < spec.w; x++ {
				dx, dy := x-cx, y-cy
				d2 := dx*dx + dy*dy
				if d2 <= rOut*rOut && d2 >= rIn*rIn {
					rgba.SetRGBA(x, y, spec.acc)
				} else if d2 < rIn*rIn {
					rgba.SetRGBA(x, y, spec.fill)
				}
			}
		}
	case "sheet":
		frameW := spec.h
		frames := spec.w / frameW
		if frames < 1 {
			frames = 1
		}
		for f := 0; f < frames; f++ {
			drawChibiFallback(rgba, f*frameW, 0, frameW, spec.h, spec.fill, spec.acc, outline, f%2)
		}
	case "stop":
		drawBoxFallback(rgba, 4, 6, spec.w-8, spec.h-10, spec.fill, spec.acc, outline)
	case "bus":
		drawBoxFallback(rgba, 2, 6, spec.w-4, spec.h-12, spec.fill, spec.acc, outline)
	default:
		drawChibiFallback(rgba, 0, 0, spec.w, spec.h, spec.fill, spec.acc, outline, 0)
	}

	return ebiten.NewImageFromImage(rgba)
}

func drawChibiFallback(img *image.RGBA, ox, oy, w, h int, body, headAcc, outline color.RGBA, bob int) {
	cx := ox + w/2
	headTop := oy + 3 + bob
	headH := h * 52 / 100
	headW := w * 50 / 100
	bodyTop := headTop + headH
	bodyBot := oy + h - 3

	for y := headTop; y < bodyBot; y++ {
		for x := ox + 2; x < ox+w-2; x++ {
			dx := x - cx
			if dx < 0 {
				dx = -dx
			}
			var halfW int
			var c color.RGBA
			if y < bodyTop {
				halfW = headW / 2
				c = headAcc
				if y < headTop+4 {
					c = body
				}
			} else {
				halfW = headW/2 - 2
				c = body
			}
			if dx == halfW || y == headTop || y == bodyBot-1 {
				img.SetRGBA(x, y, outline)
			} else if dx < halfW {
				img.SetRGBA(x, y, c)
			}
		}
	}
}

func drawBoxFallback(img *image.RGBA, x0, y0, w, h int, fill, acc, outline color.RGBA) {
	for y := y0; y < y0+h; y++ {
		for x := x0; x < x0+w; x++ {
			if x == x0 || x == x0+w-1 || y == y0 || y == y0+h-1 {
				img.SetRGBA(x, y, outline)
			} else if y < y0+h/3 {
				img.SetRGBA(x, y, acc)
			} else {
				img.SetRGBA(x, y, fill)
			}
		}
	}
}
