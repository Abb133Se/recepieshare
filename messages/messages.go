package messages

import (
	"fmt"
	"strings"
)

var translations = map[string]map[string]string{
	"en": {
		// Common
		"SUCCESS":           "Operation completed successfully",
		"Failed":            "Operation failed",
		"Unauthorized":      "Unauthorized",
		"Forbidden":         "Forbidden",
		"BadRequest":        "Bad Request",
		"InternalServerErr": "Internal server error",
		"NotFound":          "Not found",
		"DBConnectionErr":   "Dtabase connection error",

		// Recipe
		"RecipeNotFound":             "Recipe not found",
		"RecipeCreated":              "Recipe created successfully",
		"RecipeCreateFailed":         "Failed to create recipe",
		"RecipeUpdated":              "Recipe updated successfully",
		"RecipeDeleted":              "Recipe deleted successfully",
		"RecipeDeleteFail":           "Failed to delete recipe",
		"RecipeDeleteForbidden":      "Recipe delete forbidden",
		"RecipeHasNoIngredient":      "Recipe has no igrediens",
		"RecipeIngredientsOk":        "Ingredients fetched succussfully",
		"RecipeFetchFail":            "Failed to fetch recipe",
		"RecipeIngredientsFetchFail": "Faled to fetch recipe ingredients",
		"RecipeUpdateFail":           "Failed to update recipe",
		"RecipeNutritionFail":        "Failed to fetch recipe nutritional valuse",
		"RecipeTagFetchFail":         "Failed to fetch recipe tags",
		"RecipeTagCreateFail":        "Failed to create tag(s)",
		"RecipeTagQueryFail":         "Failed to query tags(s)",
		"RecipeTagUpdateFail":        "Failed to update recipe tag(s)",
		"RecipeTagUpdated":           "Recipe tags updated",
		"RecipeTagDeleteFail":        "Failed to delete recipe tag(s)",
		"RecipeTagDeleted":           "recipe tag(s) deleted succussfully",
		"RecipeCatsFetcFail":         "Failed to fetch recipe categories",
		"RecipeCatsDeleteFail":       "Faild to delete recipe categories",
		"RecipeCatsDeleted":          "Recipe categories deleted successfully",

		// User
		"LoginInvalidEmailPass":     "Invalid email or password",
		"UserAlreadyExists":         "User already exists",
		"UserCreatedSuccess":        "User created successfully",
		"UserCreateFailed":          "Failed to create user",
		"UserNotFound":              "User not found",
		"PasswordResetSent":         "Passwor reset link sent to email",
		"PasswordResetFailed":       "failed to send reset password link",
		"PasswordResetSuccess":      "Password reset successfully",
		"TokenExpired":              "Token is expired",
		"TokenInvalid":              "Token is invalid",
		"EmptyInfoErr":              "User information shouldn't be empty",
		"EmailCheckErr":             "Failed to check email",
		"EmailExistsErr":            "Email ALready exists",
		"UserFetchFail":             "Failed to trtrieve user",
		"GeneratTokenFail":          "Failed to generae token",
		"GenerateSaltFail":          "Failed to generate Salt",
		"PasswordResetCreateFailed": "Failed to store reset token",
		"UserStatFetchFail":         "Failed to fetch user stats",

		// DB
		"DB_ERROR": "Database error",
		"DB_SAVE":  "Failed to save record",
		"DB_CONN":  "Database connection failed",

		// Comment
		"CommentNotFound":        "Comment not found",
		"CommentPosted":          "Comment posted successfully",
		"CommentDeleted":         "Comment deleted successfully",
		"CommentAlreadyExists":   "Comment already exists",
		"CommentDeleteForbidden": "Comment deletion forbidden",
		"CommentLikeSuccess":     "Comment liked successfully",
		"CommentDislikeSuccess":  "Comment diliked successfully",
		"CommnetFetchFail":       "Failed to fetch comment",
		"CommentPostFail":        "Failed to post comment",
		"CommentPost":            "Posted comment successfully",
		"CommentDeleteFail":      "Failed to delete comment",
		"CommentDislikeFail":     "Failed to dislike comment",
		"CommentLikeFail":        "Failed to like comment",

		// Favorite
		"FavoriteAdded":           "Added to favorites",
		"FavoriteRemoved":         "Removed from favorites",
		"FavoriteNotFound":        "Removed not found",
		"FavoriteExists":          "Favorite already exists",
		"FavoriteFailed":          "Fialed to do operation on favorite",
		"FavoriteAddFail":         "Failed to add favorite",
		"FavoriteRemoveQueryFail": "Failed to query favorites",
		"FavoriteRemoveFail":      "Failed to remove favorite",

		// Rating
		"RatingNotFound":        "Rating not found",
		"RatingAdded":           "Rating submitted",
		"RatingUpdated":         "Rating updated",
		"RatingDeleted":         "Rating deleted",
		"RatingInvalidScore":    "Invalid rating score",
		"RatingUpdateFailed":    "Failed to update rating",
		"RatingDeleteForbidden": "Rating deletion forbidden",
		"RatingAddFail":         "Failed to add rating",
		"RatingDeleteFail":      "Failed to delete rating",
		"RatingFetchFail":       "Failed to fetch rating",

		// Image
		"ImageUploaded":        "Image uploaded successfully",
		"ImageDeleted":         "Image deleted successfully",
		"ImageNotFound":        "Image not found",
		"ImageUploadForbidden": "Image uploaded forbidden",
		"ImageDeleteForbidden": "Image delete forbidden",
		"ImageServeFailed":     "Failed to load the image",

		// Category
		"CatCreationOk":               "Category created succussful",
		"CatCreationFailed":           "Failed to create catifry",
		"CatAlreadyExists":            "Category Already exists",
		"CatUploadOK":                 "Category uploaded successfully",
		"CatUploadFailed":             "Failed to update category",
		"CatDeletionOk":               "Category deleted successfully",
		"CatDeletionFaied":            "Failed to delete category",
		"CatFetchFailed":              "Failed to fetch category",
		"CatFailedAssocioationRemova": "Failed to delete category associations",
		"CatNotFound":                 "Category Not Found",
		"CatUpdateFail":               "Failed to update category",
		"CatUpdateOk":                 "Category updated successfully",

		// Tag
		"TagCreationOk":               "Tag created succussful",
		"TagCreationFailed":           "Failed to create catifry",
		"TagAlreadyExists":            "Tag Already exists",
		"TagUploadOK":                 "Tag uploaded successfully",
		"TagUploadFailed":             "Failed to update tag",
		"TagDeletionOk":               "Tagegory deleted successfully",
		"TagDeletionFaied":            "failed to delete tag",
		"TagFetchFailed":              "Failed to fetch tag",
		"TagFailedAssocioationRemova": "Failed to delete tag associations",
		"TagNotFound":                 "Tag Not Found",
	},

	"fa": {
		// Common
		"SUCCESS":           "عملیات با موفقیت انجام شد",
		"Failed":            "عملیات ناموفق بود",
		"Unauthorized":      "دسترسی غیرمجاز",
		"Forbidden":         "اجازه دسترسی ندارید",
		"BadRequest":        "درخواست نامعتبر",
		"InternalServerErr": "خطای داخلی سرور",
		"NotFound":          "یافت نشد",
		"DBConnectionErr":   "خطا در اتصال به پایگاه داده",

		// Recipe
		"RecipeNotFound":             "دستور پخت یافت نشد",
		"RecipeCreated":              "دستور پخت با موفقیت ساخته شد",
		"RecipeCreateFailed":         "ساخت دستور پخت با خطا مواجه شد",
		"RecipeUpdated":              "دستور پخت با موفقیت بروزرسانی شد",
		"RecipeDeleted":              "دستوز پخت با موفقیت حذف شد",
		"RecipeDeleteFail":           "خطا در حذف دستور پخت",
		"RecipeDeleteForbidden":      "اجازه حذف این دستور را ندارید",
		"RecipeHasNoIngredient":      "دستور پخت هیچ ماده اولیه ای ندارد",
		"RecipeIngredientsOk":        "مواد اولیه با موفقیت دریافت شد",
		"RecipeFetchFail":            "خطا در درسافت دستور پخت",
		"RecipeIngredientsFetchFail": "خطا در باگذاری مواد اولیه دستور پخت",
		"RecipeUpdateFail":           "خطا در بروزرسانی دستورپخت",
		"RecipeNutritionFail":        "خطا در دریاف اطلاعات ارزش غذایی",
		"RecipeTagFetchFail":         "خطا در دریافت برچسب های دستور پخت",
		"RecipeTagCreateFail":        "خطا در ساخت برچسب ها",
		"RecipeTagQueryFail":         "خطا در ارسال برچسب ها",
		"RecipeTagUpdateFail":        "خطا در بروزرسانی برچسب ها",
		"RecipeTagUpdated":           "با موفقیت بروزرسانی شد",
		"RecipeTagDeleteFail":        "خطا در حذف برچسب",
		"RecipeTagDeleted":           "حذف برچسب با موفقیت انجام شد",
		"RecipeCatsFetcFail":         "خطا در دریافت دسته بندی ها",
		"RecipeCatsDeleteFail":       "خطا در حذف دسته بندی ها",
		"RecipeCatsDeleted":          "دسته بندی ها با موفقیت حذف شد",

		// User
		"LoginInvalidEmailPass":     "ایمیل یا رمز عبور نامعتبر است",
		"UserAlreadyExists":         "کاربر با این ایمیل قبلا ثبت شده",
		"UserCreatedSuccess":        "کاربر با موفقیت ثبت شد",
		"UserCreateFailed":          "خطا در ایجاد کاربر",
		"UserNotFound":              "کاربر یافت نشد",
		"PasswordResetSent":         "درصورت وجود ایمیل لینک بازیابی ارسال شد",
		"PasswordResetFailed":       "خطا در ارسال لینک بازیابی",
		"PasswordResetSuccess":      "رمز عبور با موفقیت تغییر کرد",
		"TokenExpired":              "توکن منقضی شده است",
		"TokenInvalid":              "توکن نامعتبر است",
		"EmptyInfoErr":              "اطلاعات کاربری نباید خالی باشند",
		"EmailCheckErr":             "خطا در بررسی ایمیل",
		"EmailExistsErr":            "این ایمیل قبلا در سایت ثبت شده است",
		"UserFetchFail":             "خطا در دریافت اطلاعات کاربر",
		"GeneratTokenFail":          "خطا در ساخت توکن",
		"GenerateSaltFail":          "خطا در ساخت سالت",
		"PasswordResetCreateFailed": "خطا در ذخیره سازی توکن بازیابی",
		"UserStatFetchFail":         "خطا در بارگذاری اطلاعات کاربر",

		// DB
		"DB_ERROR": "خطای پایگاه داده",
		"DB_SAVE":  "ذخیره رکورد با شکست مواجه شد",
		"DB_CONN":  "اتصال به پایگاه داده برقرار نشد",

		// Comment
		"CommentNotFound":        "نظر یافت نشد",
		"CommentPosted":          "نظر با موفقیت اضافه شد",
		"CommentDeleted":         "نظر با موفقیت حذف شد",
		"CommentAlreadyExists":   "شما قبلا به این دستور نظر داده اید",
		"CommentDeleteForbidden": "شا مجاز به حذف این تسور نیستید",
		"CommentLikeSuccess":     "نظر با موفقیت لایک شد",
		"CommentDislikeSuccess":  "نظر با موفقیت دیسلایک شد",
		"CommnetFetchFail":       "خطا در بارگذاری نظر",
		"CommentPostFail":        "خطا در پست کردن نظر",
		"CommentPost":            "نظر با موفقیت پست شد",
		"CommentDeleteFail":      "نظر با موفقیت حذف شد",
		"CommentDislikeFail":     "خطا در دیسلایک کردن نظر",
		"CommentLikeFail":        "خطا در لایک کردن نظر",

		// Favorite
		"FavoriteAdded":           "به علاقه مندی ها اضافه شد",
		"FavoriteRemoved":         "از علاقه مندی ها حذف شد",
		"FavoriteNotFound":        "عالقه مندی یافت نشد",
		"FavoriteExists":          "علاقه مندیقبلا اضافه شده است",
		"FavoriteFailed":          "خطا در انجاام دستورات علاقه مندی ها",
		"FavoriteAddFail":         "خطا در اضافه کردن علاقه مندی",
		"FavoriteRemoveQueryFail": "خطا در ارسال دستور حذف",
		"FavoriteRemoveFail":      "خطا در حذف علاقه مندی",

		// Rating
		"RatingNotFound":        "امتیاز یافت نشد",
		"RatingAdded":           "امتیاز با موفقیت ثبت شد",
		"RatingUpdated":         "امتیاز بروز رسانی شد",
		"RatingDeleted":         "امتیاز با موفقیت حذف شد",
		"RatingInvalidScore":    "امتیاز باید بین 1 تا 5 باشد",
		"RatingUpdateFailed":    "خطا در بروزرسانی امتیاز",
		"RatingDeleteForbidden": "اجازه حذف این امتیاز را ندارید",
		"RatingAddFail":         "خطا در اضافه کردن امتیاز",
		"RatingDeleteFail":      "خطا در حذف امتیاز",
		"RatingFetchFail":       "خطا در دریافت امتیاز",

		// Image
		"ImageUploaded":        "تصویر با موفقیت بارگذاری شد",
		"ImageDeleted":         "تصویر با موفقیت حذف شد",
		"ImageNotFound":        "تصویر یافت نشد",
		"ImageUploadForbidden": "اجازه بارگذاری تصویر برای این مورد را ندارید",
		"ImageDeleteForbidden": "اجازه حذف این تصویر را ندارید",
		"ImageServeFailed":     "خطا در بارگذاری تصویر",

		// Category
		"CatCreationOk":               "دسته بندی با موفقیت ساخته شد",
		"CatCreationFailed":           "خطا در ساخت دسته بندی",
		"CatAlreadyExists":            "این دسته بندی موجود میباشد",
		"CatUploadOK":                 "دسته بندی با موفقیت برورسانی شد",
		"CatUploadFailed":             "خطا در بروزرسانی دسته بندی",
		"CatDeletionOk":               "دسته بندی با موفقیت حذف شد",
		"CatDeletionFaied":            "خطا در حذف دسته بندی",
		"CatFetchFailed":              "خطا در دریافت دسته بندی",
		"CatFailedAssocioationRemova": "خطا در حذف ارتباطات دسته بندی",
		"CatNotFound":                 "دسته بندی یافت نشد",
		"CatUpdateFail":               "خطا در بروز رسانی دسته بندی",
		"CatUpdateOk":                 "دسته بندی با موفقیت بروز رسانی شد",

		// Tag
		"TagCreationOk":               "برچسب با موفقیت ایجاد شد",
		"TagCreationFailed":           "خطا در ایجا برچسب",
		"TagAlreadyExists":            "برچسب موجود است",
		"TagUploadOK":                 "برچسب با موفیت بروز رسانی شد",
		"TagUploadFailed":             "خطا در بروز رسانی برچسب",
		"TagDeletionOk":               "برچسب با موفقیت حذف شد",
		"TagDeletionFaied":            "خطا در حذف برچسب",
		"TagFetchFailed":              "خطا در دریافت برچسب",
		"TagFailedAssocioationRemova": "خطا در حذف ارتباطات برچسب",
		"TagNotFound":                 "برچسب یافت نشد",
	},
}

var Common = struct {
	Success           Message
	Failed            Message
	Unauthorized      Message
	Forbidden         Message
	BadRequest        Message
	InternalServerErr Message
	NotFound          Message
	DBConnectionErr   Message
}{
	Success:           Message{"SUCCESS"},
	Failed:            Message{"Failed"},
	Unauthorized:      Message{"Unauthorized"},
	Forbidden:         Message{"Forbidden"},
	BadRequest:        Message{"BadRequest"},
	InternalServerErr: Message{"InternalServerErr"},
	NotFound:          Message{"NotFound"},
	DBConnectionErr:   Message{"DBConnectionErr"},
}

var Recipe = struct {
	RecipeNotFound             Message
	RecipeCreated              Message
	RecipeCreateFailed         Message
	RecipeUpdated              Message
	RecipeUpdateFail           Message
	RecipeDeleted              Message
	RecipeDeleteFail           Message
	RecipeDeleteForbidden      Message
	RecipeHasNoIngredient      Message
	RecipeIngredientsOk        Message
	RecipeFetchFail            Message
	RecipeIngredientsFetchFail Message
	RecipeTagFetchFail         Message
	RecipeTagCreateFail        Message
	RecipeTagQueryFail         Message
	RecipeTagUpdateFail        Message
	RecipeTagUpdated           Message
	RecipeNutritionFail        Message
	RecipeTagsDeleteFail       Message
	ReciepTagsDeleted          Message
	RecipeCatsFetcFail         Message
	RecipeCatsDeleteFail       Message
	RecipeCatsDeleted          Message
}{
	RecipeNotFound:             Message{"RecipeNotFound"},
	RecipeCreated:              Message{"RecipeCreated"},
	RecipeCreateFailed:         Message{"RecipeCreateFailed"},
	RecipeUpdated:              Message{"RecipeUpdated"},
	RecipeDeleted:              Message{"RecipeDeleted"},
	RecipeDeleteFail:           Message{"RecipeDeleteFail"},
	RecipeDeleteForbidden:      Message{"RecipeDeleteForbidden"},
	RecipeHasNoIngredient:      Message{"RecipeHasNoIngredient"},
	RecipeIngredientsOk:        Message{"RecipeIngredientsOk"},
	RecipeFetchFail:            Message{"RecipeFetchFail"},
	RecipeIngredientsFetchFail: Message{"RecipeIngredientsFetchFail"},
	RecipeUpdateFail:           Message{"RecipeUpdateFail"},
	RecipeNutritionFail:        Message{"RecipeNutritionFail"},
	RecipeTagFetchFail:         Message{"RecipeTagFetchFail"},
	RecipeTagCreateFail:        Message{"RecipeTagCreateFail"},
	RecipeTagQueryFail:         Message{"RecipeTagQueryFail"},
	RecipeTagUpdateFail:        Message{"RecipeTagUpdateFail"},
	RecipeTagUpdated:           Message{"RecipeTagupdated"},
	RecipeTagsDeleteFail:       Message{"RecipeTagDeleteFail"},
	ReciepTagsDeleted:          Message{"RecipeTagDeleted"},
	RecipeCatsFetcFail:         Message{"RecipeCatsFetcFail"},
	RecipeCatsDeleteFail:       Message{"RecipeCatsDeleteFail"},
	RecipeCatsDeleted:          Message{"RecipeCatsDeleted"},
}

var User = struct {
	LoginInvalidEmailPass     Message
	UserAlreadyExists         Message
	UserCreatedSuccess        Message
	UserCreateFailed          Message
	UserFetchFail             Message
	UserNotFound              Message
	GeneratTokenFail          Message
	GenerateSaltFail          Message
	PasswordResetSent         Message
	PasswordResetFailed       Message
	PasswordResetCreateFailed Message
	PasswordResetSuccess      Message
	TokenExpired              Message
	TokenInvalid              Message
	EmptyInfoErr              Message
	EmailCheckErr             Message
	EmailExistsErr            Message
	UserStatFetchFail         Message
}{
	LoginInvalidEmailPass:     Message{"LoginInvalidEmailPass"},
	UserAlreadyExists:         Message{"UserAlreadyExists"},
	UserCreatedSuccess:        Message{"UserCreatedSuccess"},
	UserCreateFailed:          Message{"UserCreateFailed"},
	UserNotFound:              Message{"UserNotFound"},
	PasswordResetSent:         Message{"PasswordResetSent"},
	PasswordResetFailed:       Message{"PasswordResetFailed"},
	PasswordResetSuccess:      Message{"PasswordResetSuccess"},
	TokenExpired:              Message{"TokenExpired"},
	TokenInvalid:              Message{"TokenInvalid"},
	EmptyInfoErr:              Message{"EmptyInfoErr"},
	EmailCheckErr:             Message{"EmailCheckErr"},
	EmailExistsErr:            Message{"EmailExistsErr"},
	UserFetchFail:             Message{"UserFetchFail"},
	GeneratTokenFail:          Message{"GeneratTokenFail"},
	GenerateSaltFail:          Message{"GenerateSaltFail"},
	PasswordResetCreateFailed: Message{"PasswordResetCreateFailed"},
	UserStatFetchFail:         Message{"UserStatFetchFail"},
}

var Comment = struct {
	CommentNotFound        Message
	CommentPost            Message
	CommentPostFail        Message
	CommentPosted          Message
	CommentDeleted         Message
	CommentAlreadyExists   Message
	CommentDeleteForbidden Message
	CommentDeleteFail      Message
	CommentLikeSuccess     Message
	CommentLikeFail        Message
	CommentDislikeSuccess  Message
	CommentDislikeFail     Message
	CommnetFetchFail       Message
}{
	CommentNotFound:        Message{"CommentNotFound"},
	CommentPosted:          Message{"CommentPosted"},
	CommentDeleted:         Message{"CommentDeleted"},
	CommentAlreadyExists:   Message{"CommentAlreadyExists"},
	CommentDeleteForbidden: Message{"CommentDeleteForbidden"},
	CommentLikeSuccess:     Message{"CommentLikeSuccess"},
	CommentDislikeSuccess:  Message{"CommentDislikeSuccess"},
	CommnetFetchFail:       Message{"CommnetFetchFail"},
	CommentPostFail:        Message{"CommentPostFail"},
	CommentPost:            Message{"CommentPost"},
	CommentDeleteFail:      Message{"CommentDeleteFail"},
	CommentLikeFail:        Message{"CommentLikeFail"},
	CommentDislikeFail:     Message{"CommentDislikeFail"},
}

var Favorite = struct {
	FavoriteAdded           Message
	FavoriteRemoved         Message
	FavoriteRemoveFail      Message
	FavoriteRemoveQueryFail Message
	FavoriteNotFound        Message
	FavoriteExists          Message
	FavoriteFailed          Message
	FavoriteAddFail         Message
}{
	FavoriteAdded:           Message{"FavoriteAdded"},
	FavoriteRemoved:         Message{"FavoriteRemoved"},
	FavoriteNotFound:        Message{"FavoriteNotFound"},
	FavoriteExists:          Message{"FavoriteExists"},
	FavoriteFailed:          Message{"FavoriteFailed"},
	FavoriteAddFail:         Message{"FavoriteAddFail"},
	FavoriteRemoveQueryFail: Message{"FavoriteRemoveQueryFail"},
	FavoriteRemoveFail:      Message{"FavoriteRemoveFail"},
}

var Rating = struct {
	RatingNotFound        Message
	RatingAdded           Message
	RatingAddFail         Message
	RatingUpdated         Message
	RatingDeleted         Message
	RatingInvalidScore    Message
	RatingUpdateFailed    Message
	RatingDeleteForbidden Message
	RatingFetchFail       Message
	RatingDeleteFail      Message
}{
	RatingNotFound:        Message{"RatingNotFound"},
	RatingAdded:           Message{"RatingAdded"},
	RatingUpdated:         Message{"RatingUpdated"},
	RatingDeleted:         Message{"RatingDeleted"},
	RatingInvalidScore:    Message{"RatingInvalidScore"},
	RatingUpdateFailed:    Message{"RatingUpdateFailed"},
	RatingDeleteForbidden: Message{"RatingDeleteForbidden"},
	RatingAddFail:         Message{"RatingAddFail"},
	RatingFetchFail:       Message{"RatingFetchFail"},
	RatingDeleteFail:      Message{"RatingDeleteFail"},
}

var Image = struct {
	ImageUploaded        Message
	ImageDeleted         Message
	ImageNotFound        Message
	ImageUploadForbidden Message
	ImageDeleteForbidden Message
	ImageServeFailed     Message
}{
	ImageUploaded:        Message{"ImageUploaded"},
	ImageDeleted:         Message{"ImageDeleted"},
	ImageNotFound:        Message{"ImageNotFound"},
	ImageUploadForbidden: Message{"ImageUploadForbidden"},
	ImageDeleteForbidden: Message{"ImageDeleteForbidden"},
	ImageServeFailed:     Message{"ImageServeFailed"},
}

var Category = struct {
	CatCreationOk               Message
	CatCreationFailed           Message
	CatAlreadyExists            Message
	CatUploadOK                 Message
	CatUploadFailed             Message
	CatDeletionOk               Message
	CatDeletionFaied            Message
	CatFetchFailed              Message
	CatFailedAssocioationRemova Message
	CatNotFound                 Message
	CatUpdateOk                 Message
	CatUpdateFail               Message
}{
	CatCreationOk:               Message{"CatCreationOk"},
	CatCreationFailed:           Message{"CatCreationFailed"},
	CatAlreadyExists:            Message{"CatAlreadyExists"},
	CatUploadOK:                 Message{"CatUploadOK"},
	CatUploadFailed:             Message{"CatUploadFailed"},
	CatDeletionOk:               Message{"CatDeletionOk"},
	CatDeletionFaied:            Message{"CatDeletionFaied"},
	CatFetchFailed:              Message{"CatFetchFailed"},
	CatFailedAssocioationRemova: Message{"CatFailedAssocioationRemova"},
	CatNotFound:                 Message{"CatNotFound"},
	CatUpdateOk:                 Message{"CatUpdateOk"},
	CatUpdateFail:               Message{"CatUpdateFail"},
}

var Tag = struct {
	TagNotFound                 Message
	TagCreationOk               Message
	TagCreationFailed           Message
	TagAlreadyExists            Message
	TagUploadOK                 Message
	TagUploadFailed             Message
	TagDeletionOk               Message
	TagDeletionFaied            Message
	TagFetchFailed              Message
	TagFailedAssocioationRemova Message
}{
	TagNotFound:                 Message{"TagNotFound"},
	TagCreationOk:               Message{"TagCreationOk"},
	TagCreationFailed:           Message{"TagCreationFailed"},
	TagAlreadyExists:            Message{"TagAlreadyExists"},
	TagUploadOK:                 Message{"TagUploadOK"},
	TagUploadFailed:             Message{"TagUploadFailed"},
	TagDeletionOk:               Message{"TagDeletionOk"},
	TagDeletionFaied:            Message{"TagDeletionFaied"},
	TagFetchFailed:              Message{"TagFetchFailed"},
	TagFailedAssocioationRemova: Message{"TagFailedAssocioationRemova"},
}

var defaultLang = "en"
var CurrentLang = defaultLang

func SetLang(lang string) {
	lang = strings.ToLower(lang)
	if _, ok := translations[lang]; ok {
		CurrentLang = lang
	} else {
		CurrentLang = defaultLang
	}
}

func T(lang, key string) string {
	if vals, ok := translations[lang]; ok {
		if msg, exists := vals[key]; exists {
			return msg
		}
	}
	return key
}

type Message struct {
	Key string
}

func (m Message) String() string {
	fmt.Println(CurrentLang, m.Key)
	return T(CurrentLang, m.Key)
}
