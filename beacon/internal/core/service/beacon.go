package service

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"strings"
	"time"

	"beacon/config"
	"beacon/internal/core/entity"
	"beacon/internal/core/ports"
)

type BeaconService struct {
	notifier     ports.Notifier
	provider     ports.ChallengeProvider
	reminderType config.ReminderType
}

func NewBeaconService(
	notifier ports.Notifier,
	provider ports.ChallengeProvider,
	reminderType config.ReminderType,
) *BeaconService {
	return &BeaconService{
		notifier:     notifier,
		provider:     provider,
		reminderType: reminderType,
	}
}

func (s *BeaconService) Run(ctx context.Context) error {
	msg, err := s.buildMessage(ctx)
	if err != nil {
		return fmt.Errorf("build %s message: %w", s.reminderType, err)
	}

	log.Printf("📨 Sending %s reminder...", s.reminderType)

	if err := s.notifier.Send(ctx, msg); err != nil {
		return fmt.Errorf("send %s reminder: %w", s.reminderType, err)
	}

	log.Printf("✅ %s reminder sent", s.reminderType)
	return nil
}

func (s *BeaconService) buildMessage(ctx context.Context) (entity.Message, error) {
	switch s.reminderType {
	case config.ReminderHealthy:
		return s.buildHealthyMessage(), nil
	case config.ReminderReview:
		return s.buildReviewMessage(), nil
	default:
		return s.buildDSAMessage(ctx)
	}
}

func (s *BeaconService) buildDSAMessage(ctx context.Context) (entity.Message, error) {
	challenge, err := s.provider.FetchDaily(ctx)
	if err != nil {
		return entity.Message{}, fmt.Errorf("fetch daily challenge: %w", err)
	}

	diffEmoji := map[string]string{
		"Easy":   "🟢",
		"Medium": "🟡",
		"Hard":   "🔴",
	}

	emoji := diffEmoji[challenge.Difficulty]
	if emoji == "" {
		emoji = "⚪"
	}

	tags := "—"
	if len(challenge.TopicTags) > 0 {
		tags = strings.Join(challenge.TopicTags, ", ")
	}

	description := fmt.Sprintf(
		"**#%s. %s**\n\n"+
			"%s **Difficulty:** %s\n"+
			"🏷️ **Topics:** %s\n\n"+
			"**[→ Solve Now](%s)**",
		challenge.ID,
		challenge.Title,
		emoji,
		challenge.Difficulty,
		tags,
		challenge.Link,
	)

	greetings := []string{
		"🧠 Daily Challenge đã sẵn sàng!",
		"⚔️ Bài hôm nay đang chờ bạn!",
		"🔥 Một ngày không luyện là một ngày lùi!",
		"💪 Consistency beats talent. Let's go!",
		"🎯 Big Tech không tự đến. Solve it!",
	}

	return entity.Message{
		Embeds: []entity.Embed{
			{
				Title:       pick(greetings),
				Description: description,
				Color:       0xFFA116,
			},
		},
	}, nil
}

var (
	waterTips = []string{
		"💧 Uống một ly nước đi, cơ thể cần hydrate!",
		"🥤 Đã lâu rồi chưa uống nước, nạp ngay một ly!",
		"💦 Não cần nước để hoạt động hiệu quả, uống ngay!",
		"🫗 Nước lọc > cà phê. Uống 1 ly nước nào!",
	}

	standTips = []string{
		"🚶 Đứng dậy đi lại vài bước cho máu lưu thông!",
		"🧍 Ngồi lâu quá rồi, đứng lên vươn vai nào!",
		"🏃 Đi lại 2-3 phút cho cơ thể bớt cứng!",
		"🦵 Đứng lên, duỗi chân, ngồi lại tư thế đúng!",
	}

	eyeTips = []string{
		"👀 Nhìn xa 20 giây để mắt nghỉ ngơi (20-20-20)!",
		"🧘 Nhắm mắt 20 giây, xoa nhẹ vùng thái dương!",
		"👁️ Massage mắt nhẹ nhàng theo vòng tròn 10 giây!",
		"😌 Chớp mắt liên tục 15 lần cho mắt bớt khô!",
	}

	stretchTips = []string{
		"🙆 Xoay cổ, kéo vai, giãn lưng cho đỡ mỏi!",
		"💆 Nghiêng đầu sang hai bên, giãn cơ cổ nào!",
		"🤸 Vặn người sang trái-phải, giải phóng lưng!",
		"✋ Xoay cổ tay, duỗi ngón — bàn tay cũng cần nghỉ!",
	}

	postureTips = []string{
		"🪑 Check tư thế ngồi: lưng thẳng, vai thả lỏng!",
		"📐 Màn hình ngang tầm mắt, khuỷu tay 90 độ!",
		"🧱 Lưng tựa vào ghế, chân chạm sàn, vai relaxed!",
	}

	breathTips = []string{
		"🌬️ Hít sâu 4 giây — giữ 4 giây — thở ra 4 giây!",
		"🧘 3 hơi thở sâu để reset lại tinh thần!",
		"💨 Box breathing: hít 4s → giữ 4s → thở 4s → giữ 4s!",
	}
)

func (s *BeaconService) buildHealthyMessage() entity.Message {
	allCategories := [][]string{waterTips, standTips, eyeTips, stretchTips, postureTips, breathTips}
	rand.Shuffle(len(allCategories), func(i, j int) {
		allCategories[i], allCategories[j] = allCategories[j], allCategories[i]
	})

	count := 3 + rand.IntN(2)
	if count > len(allCategories) {
		count = len(allCategories)
	}

	var parts []string
	for i := 0; i < count; i++ {
		parts = append(parts, pick(allCategories[i]))
	}

	return entity.Message{
		Embeds: []entity.Embed{
			{
				Title:       "⏰ Health Break!",
				Description: strings.Join(parts, "\n"),
				Color:       0x2ECC71,
			},
		},
	}
}

func (s *BeaconService) buildReviewMessage() entity.Message {
	now := time.Now().In(time.FixedZone("UTC+7", 7*3600))
	weekday := vietnameseWeekday(now.Weekday())
	dateStr := now.Format("02/01/2006")

	prompts := []string{
		fmt.Sprintf("📅 **%s, %s**\n\nHôm nay kết thúc rồi. Bạn có tốt hơn hôm qua 1%% không?\n\nMỗi ngày 1%% — sau 1 năm bạn sẽ giỏi hơn 37 lần.\nĐừng so sánh với người khác, hãy so sánh với chính mình ngày hôm qua.\n\n**Ngày mai, hãy làm tốt hơn hôm nay.**", weekday, dateStr),
		fmt.Sprintf("📅 **%s, %s**\n\nMột ngày nữa đã qua. Bạn đã cố gắng chưa?\n\nKhông cần hoàn hảo, chỉ cần tiến bộ.\nNhững người thành công không phải giỏi nhất — họ là người **không bao giờ dừng lại**.\n\n**Keep going. Bạn đang trên đúng đường.**", weekday, dateStr),
		fmt.Sprintf("📅 **%s, %s**\n\n9h tối rồi. Trước khi nghỉ — nhìn lại một chút.\n\nBạn đã invest vào bản thân hôm nay chưa?\nMỗi giờ bạn bỏ ra hôm nay sẽ trả lại gấp bội trong tương lai.\n\n**Discipline = Freedom. Đầu tư vào bản thân là khoản đầu tư tốt nhất.**", weekday, dateStr),
		fmt.Sprintf("📅 **%s, %s**\n\nKhông ai thành công trong một đêm.\nNhưng mỗi đêm, bạn có thể tự hỏi: mình đã đi được bao xa?\n\nHãy tự hào về những gì bạn đã làm hôm nay, dù nhỏ.\n**Rồi ngày mai, đi thêm một bước nữa.**", weekday, dateStr),
		fmt.Sprintf("📅 **%s, %s**\n\nCompound effect: những việc nhỏ lặp lại mỗi ngày tạo ra kết quả lớn.\n\nBạn có maintain streak hôm nay không?\nKhông sao nếu chưa — quan trọng là **ngày mai bắt đầu lại**.\n\n**Your future self will thank you.**", weekday, dateStr),
	}

	return entity.Message{
		Embeds: []entity.Embed{
			{
				Title:       "🌙 Daily Review — Hôm nay thế nào?",
				Description: pick(prompts),
				Color:       0x9B59B6,
			},
		},
	}
}

func vietnameseWeekday(w time.Weekday) string {
	days := map[time.Weekday]string{
		time.Monday:    "Thứ Hai",
		time.Tuesday:   "Thứ Ba",
		time.Wednesday: "Thứ Tư",
		time.Thursday:  "Thứ Năm",
		time.Friday:    "Thứ Sáu",
		time.Saturday:  "Thứ Bảy",
		time.Sunday:    "Chủ Nhật",
	}
	return days[w]
}

func pick(items []string) string {
	return items[rand.IntN(len(items))]
}
