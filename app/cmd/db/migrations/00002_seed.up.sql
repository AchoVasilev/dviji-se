INSERT INTO categories(id, name, image_url)
VALUES ('5206274b-b473-49c1-bcca-6516838a9f1e', 'Рецепти',
        'https://res.cloudinary.com/dpo3vbxnl/image/upload/v1691940539/onlygains/categories/istock-1155240408_stpuam.jpg'),
       ('a1f06571-2f00-43a8-9ab9-401c802237cd', 'Упражнения',
        'https://res.cloudinary.com/dpo3vbxnl/image/upload/v1691940754/onlygains/categories/shutterstock_635024150opt_featured_1614948744_1200x672_acf_cropped_a40vos.jpg'),
       ('9fd30320-8e71-4a76-aff8-8b82fccb68e7', 'Тренировъчни програми',
        'https://res.cloudinary.com/dpo3vbxnl/image/upload/v1691941043/onlygains/categories/TheNotebook_pd92nm.jpg'),
       ('c40af472-04fc-4c51-aeb6-22818645f824', 'Фитнес зали',
        'https://res.cloudinary.com/dpo3vbxnl/image/upload/v1691941224/onlygains/categories/brutally-hardcore-gyms-you-need-to-train-at-before-you-die-652x400-10-1496399800_ahp7xa.jpg'),
       ('5e01b7ad-c631-470e-b330-54ac9a503b57', 'Хранителни режими',
        'https://res.cloudinary.com/dpo3vbxnl/image/upload/v1691941490/onlygains/categories/image04-4_hgly6c.jpg'),
       ('272e0edd-11e8-4709-bbba-5d36d77dd428', 'Сред природата',
        'https://res.cloudinary.com/dpo3vbxnl/image/upload/v1691942376/onlygains/categories/hiking-trail-names_fgpox2.jpg');

INSERT INTO roles(id, name)
VALUES ('272e0edd-11e8-4709-bbba-5d36d77dd428', 'ADMIN'),
        ('272e0edd-11e8-4709-bbba-5d36d77dd315', 'USER'),
        ('312e0edd-11e8-4709-bbba-5d36d77dd455', 'MODERATOR');

INSERT INTO permissions(id, name)
VALUES  ('26eb876b-f959-45fa-95ca-65d98b1d14a0', 'comment:write'),
        ('272e0edd-11e8-4709-bbba-5d36d77dd315', 'comment:create'),
        ('612e0edd-11e8-4709-bbba-5d36d77dd136', 'post:write'),
        ('443e0edd-11e8-4709-bbba-5d36d77dd654', 'post:create'),
        ('554e0edd-11e8-4709-bbba-5d36d77dd432', 'permission:add'),
        ('123e0edd-11e8-4709-bbba-5d36d77dd987', 'permission:remove');

INSERT INTO roles_permissions(role_id, permission_id)
VALUES ('272e0edd-11e8-4709-bbba-5d36d77dd428', '26eb876b-f959-45fa-95ca-65d98b1d14a0'),
        ('272e0edd-11e8-4709-bbba-5d36d77dd428', '272e0edd-11e8-4709-bbba-5d36d77dd315'),
        ('272e0edd-11e8-4709-bbba-5d36d77dd428', '612e0edd-11e8-4709-bbba-5d36d77dd136'),
        ('272e0edd-11e8-4709-bbba-5d36d77dd428', '443e0edd-11e8-4709-bbba-5d36d77dd654'),
        ('272e0edd-11e8-4709-bbba-5d36d77dd428', '554e0edd-11e8-4709-bbba-5d36d77dd432'),
        ('272e0edd-11e8-4709-bbba-5d36d77dd428', '123e0edd-11e8-4709-bbba-5d36d77dd987'),
        ('312e0edd-11e8-4709-bbba-5d36d77dd455', '26eb876b-f959-45fa-95ca-65d98b1d14a0'),
        ('312e0edd-11e8-4709-bbba-5d36d77dd455', '272e0edd-11e8-4709-bbba-5d36d77dd315'),
        ('312e0edd-11e8-4709-bbba-5d36d77dd455', '612e0edd-11e8-4709-bbba-5d36d77dd136'),
        ('312e0edd-11e8-4709-bbba-5d36d77dd455', '443e0edd-11e8-4709-bbba-5d36d77dd654'),
        ('272e0edd-11e8-4709-bbba-5d36d77dd315', '26eb876b-f959-45fa-95ca-65d98b1d14a0'),
        ('272e0edd-11e8-4709-bbba-5d36d77dd315', '272e0edd-11e8-4709-bbba-5d36d77dd315');
